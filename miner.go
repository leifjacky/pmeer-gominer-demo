package main

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	BigOne       = new(big.Int).SetInt64(1)
	MaxExtraonce = new(big.Int)
)

type StratumMiner struct {
	cfg *StratumMinerConfig

	jobDifficulty   float64
	jobTarget       *big.Int
	extranonce1     string
	extranonce2Size int
	job             atomic.Value
	cnt             int64

	writeMu sync.Mutex
	conn    net.Conn
}

type Job struct {
	sync.Mutex
	jobId           string
	extranonce1     string
	extranonce2     string
	extranonce2Size int
	nonce           uint32

	prevHash     string
	coinbase1    string
	coinbase2    string
	coinbase3    string
	coinbase4    string
	merkleBranch []string
	version      string
	nbits        string
	ntime        string

	coinbase     string
	coinbaseHash string
	txroot       string
	headerPrefix string
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (j *Job) RefreshJob() {
	j.GetNextExtranonce2()

	coinbaseMid := Blake2bD(MustStringToHexBytes(j.coinbase2 + j.extranonce1 + j.extranonce2 + j.coinbase3))
	coinbaseHash := Blake2bD(MustStringToHexBytes(j.coinbase1 + hex.EncodeToString(coinbaseMid) + j.coinbase4))
	j.coinbaseHash = hex.EncodeToString(coinbaseHash)
	j.coinbase = j.coinbase1 + j.coinbase2 + j.extranonce1 + j.extranonce2 + j.coinbase3 + j.coinbase4

	txroot := j.coinbaseHash
	for _, branch := range j.merkleBranch {
		txroot = hex.EncodeToString(Blake2bD(MustStringToHexBytes(txroot + branch)))
	}
	j.txroot = txroot

	//headerPrefix := j.version + j.prevHash + j.txroot + FillZeroHashLen("", 64) + j.nbits + j.ntime
	headerPrefix := ReverseStringByte(j.version) + j.prevHash + j.txroot + FillZeroHashLen("", 64) + ReverseStringByte(j.nbits) + ReverseStringByte(j.ntime)
	j.headerPrefix = headerPrefix
}

func (j *Job) GetNextExtranonce2() string {
	j.Lock()
	defer j.Unlock()
	pat := fmt.Sprintf("%%0%dx", j.extranonce2Size*2)
	j.extranonce2 = fmt.Sprintf(pat, rand.Int31())
	return j.extranonce2
}

func (j *Job) GetNextNonce() uint32 {
	j.Lock()
	defer j.Unlock()
	if j.nonce == math.MaxUint32 {
		j.RefreshJob()
		j.nonce = 0
	} else {
		j.nonce++
	}
	return j.nonce
}

func NewMiner(cfg *StratumMinerConfig) *StratumMiner {
	return &StratumMiner{
		cfg:           cfg,
		jobDifficulty: 65536,
	}
}

func (m *StratumMiner) Mine() {
	gracefulShutdownChannel := make(chan os.Signal)
	signal.Notify(gracefulShutdownChannel, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-gracefulShutdownChannel
		logrus.Warningf("receive shutdown signal")
		os.Exit(0)
	}()

	sumIntv := MustParseDuration(m.cfg.SumIntv)
	logrus.Infof("hashrate sum every %v", sumIntv)
	sumTicker := time.NewTicker(sumIntv)

	go m.start()
	for {
		select {
		case <-sumTicker.C:
			cnt := m.cnt
			m.cnt -= cnt
			logrus.Warningf("hashrates: %v", GetReadableHashRateString(float64(cnt/int64((sumIntv)/time.Second))))
		}
	}
}

func (m *StratumMiner) start() {
	th := m.cfg.Threads
	if th == 0 {
		th = runtime.NumCPU()
	}
	logrus.Infof("running with %v workers", th)
	for i := 0; i < th; i++ {
		go m.startWorker(i)
	}

	logrus.Infof("connect to %v", m.cfg.Url)
	conn, err := net.Dial("tcp", m.cfg.Url)
	if err != nil {
		logrus.Fatalf("failed to connect: %v", err)
	}
	m.conn = conn
	logrus.Infof("connected")

	buf := bufio.NewReader(conn)

	if err := m.request("mining.subscribe", []interface{}{"QitmeerMiner-v1.0.0", nil}); err != nil {
		logrus.Fatalf("error subscribe: %v", err)
	}
	data, _, err := buf.ReadLine()
	if err != nil {
		logrus.Errorf("err reading: %v", err)
		return
	}
	logrus.Debugf("recv from pool: %v", string(data))
	if err := m.handleMesg(data, 1); err != nil {
		logrus.Errorf("err handle mesg: %v", err)
		return
	}
	logrus.Infof("subscribed")

	if err := m.request("mining.authorize", []string{m.cfg.Username, m.cfg.Password}); err != nil {
		logrus.Fatalf("error authorize: %v", err)
	}
	data, _, err = buf.ReadLine()
	if err != nil {
		logrus.Errorf("err reading: %v", err)
		return
	}
	logrus.Debugf("recv from pool: %v", string(data))
	if err := m.handleMesg(data, 2); err != nil {
		logrus.Errorf("err handle mesg: %v", err)
		return
	}
	logrus.Infof("authorized")

	for {
		data, _, err := buf.ReadLine()
		if err != nil {
			logrus.Errorf("err reading: %v", err)
			return
		}

		logrus.Debugf("recv from pool: %v", string(data))
		if err := m.handleMesg(data, 0); err != nil {
			logrus.Errorf("err handle mesg: %v", err)
			return
		}
	}
	logrus.Infof("disconnected")
}

func (m *StratumMiner) handleMesg(line []byte, flag int) error {
	var mesg PoolMesg
	if err := json.Unmarshal(line, &mesg); err != nil {
		return fmt.Errorf("can't decode: %v", err)
	}
	switch flag {
	case 1:
		if mesg.Error == nil {
			result := []interface{}{}
			if err := json.Unmarshal(*mesg.Result, &result); err != nil {
				return fmt.Errorf("can't decode result: %v", err)
			}
			m.extranonce2Size = int(result[2].(float64))
			MaxExtraonce = new(big.Int).Lsh(new(big.Int).SetInt64(1), uint(m.extranonce2Size*8))
			m.extranonce1 = result[1].(string)
		} else {
			info := []interface{}{}
			if err := json.Unmarshal(*mesg.Error, &info); err != nil {
				return fmt.Errorf("can't decode error: %v", err)
			}
			return fmt.Errorf("subscribe error. %v", info[1].(string))
		}
		return nil
	case 2:
		if mesg.Error != nil {
			info := []interface{}{}
			if err := json.Unmarshal(*mesg.Error, &info); err != nil {
				return fmt.Errorf("can't decode error: %v", err)
			}
			return fmt.Errorf("authorize error. %v", info[1].(string))
		}
		return nil
	}
	switch mesg.Method {
	case "mining.set_difficulty":
		params := []float64{}
		if err := json.Unmarshal(*mesg.Params, &params); err != nil {
			return fmt.Errorf("can't decode params: %v", err)
		}
		m.jobDifficulty = params[0]
		m.jobTarget = Diff2BigTarget(m.jobDifficulty)
		logrus.Infof("job difficulty set to: %f", m.jobDifficulty)
		logrus.Infof("job target set to: %064x", m.jobTarget)
	case "mining.notify":
		params := []interface{}{}
		if err := json.Unmarshal(*mesg.Params, &params); err != nil {
			return fmt.Errorf("can't decode params: %v", err)
		}
		jobId := params[0].(string)
		prevHash := params[1].(string)
		coinbase1 := params[2].(string)
		coinbase2 := params[3].(string)
		coinbase3 := params[4].(string)
		coinbase4 := params[5].(string)
		merkleBranch := func(s []interface{}) []string {
			st := make([]string, len(s))
			for i := range s {
				st[i] = s[i].(string)
			}
			return st
		}(params[6].([]interface{}))
		versionStr := params[7].(string)
		bitsStr := params[8].(string)
		timeStr := params[9].(string)
		job := &Job{
			jobId:           jobId,
			extranonce1:     m.extranonce1,
			extranonce2Size: m.extranonce2Size,
			prevHash:        prevHash,
			coinbase1:       coinbase1,
			coinbase2:       coinbase2,
			coinbase3:       coinbase3,
			coinbase4:       coinbase4,
			merkleBranch:    merkleBranch,
			version:         versionStr,
			nbits:           bitsStr,
			ntime:           timeStr,
		}
		job.RefreshJob()
		m.job.Store(job)
		logrus.Infof("new job: %v - %v - %064x", jobId, prevHash, m.jobTarget)
	default:
		result := false
		if err := json.Unmarshal(*mesg.Result, &result); err != nil {
			return fmt.Errorf("can't decode result: %v", err)
		}
		if result {
			logrus.Infof("share accepted")
		} else {
			info := []interface{}{}
			if err := json.Unmarshal(*mesg.Error, &info); err != nil {
				return fmt.Errorf("can't decode error: %v", err)
			}
			logrus.Infof("share rejected. %v", info[1].(string))
		}
	}
	return nil
}

type JsonRpcReq struct {
	Id     int64       `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

type PoolMesg struct {
	Id     *json.RawMessage `json:"id"`
	Method string           `json:"method"`
	Result *json.RawMessage `json:"result"`
	Params *json.RawMessage `json:"params"`
	Error  *json.RawMessage `json:"error"`
}

func (m *StratumMiner) request(method string, params interface{}) error {
	return m.write(&JsonRpcReq{0, method, params})
}

var lineDelimiter = []byte("\n")

func (m *StratumMiner) write(message interface{}) error {
	b, err := json.Marshal(message)
	if err != nil {
		return err
	}

	m.writeMu.Lock()
	defer m.writeMu.Unlock()

	logrus.Debugf("write to pool: %v", string(b))
	if _, err := m.conn.Write(b); err != nil {
		return err
	}

	_, err = m.conn.Write(lineDelimiter)
	return err
}

func (m *StratumMiner) loadJob() *Job {
	job := m.job.Load()
	if job == nil {
		return nil
	}
	return job.(*Job)
}

func (m *StratumMiner) startWorker(i int) {
	for {
		job := m.loadJob()
		if job == nil {
			logrus.Warningf("#%d job not ready. sleep for 5s...", i)
			time.Sleep(5 * time.Second)
			continue
		}
		nonce := job.GetNextNonce()
		b := append(MustStringToHexBytes(job.headerPrefix), UInt32LEToBytes(nonce)...)
		b = append(b, byte(m.cfg.PowType))
		headerHashReverse := HashFunc(b)
		headerHash := ReverseBytes(headerHashReverse)
		hashBig := new(big.Int).SetBytes(headerHash)

		if hashBig.Cmp(m.jobTarget) < 0 {
			logrus.Infof("share found: %s %08x", job.extranonce2, nonce)
			logrus.Tracef("coinbase: %s", job.coinbase)
			logrus.Tracef("cbhash: %s", job.coinbaseHash)
			logrus.Tracef("header : %x", b)
			logrus.Tracef("hash: %064x", headerHash)
			go func() {
				if err := m.request("mining.submit", []interface{}{m.cfg.Username, job.jobId, job.extranonce2, job.ntime, fmt.Sprintf("%08x", nonce)}); err != nil {
					logrus.Fatalf("error submit: %v", err)
				}
			}()
		}
		m.cnt++
	}
}
