package main

import (
	"io"
	"os"
	"runtime"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

type StratumMinerConfig struct {
	Url      string
	Username string
	Password string
	SumIntv  string
	Threads  int
	PowType  PowType
}

type PowType byte

const (
	BLAKE2BD PowType = iota
	CUCKAROO
	CUCKAROO29
	CUCKATOO
	X8R16
	X16R
	KECCAK256
)

var PowTypeMap = map[string]PowType{
	"blake2bd":  BLAKE2BD,
	"keccak256": KECCAK256,
}

func main() {
	var url, username, password, loglevel, logfile, algo string
	var threads int
	pflag.StringVarP(&url, "url", "o", "pmeer.uupool.cn:8667", "stratum pool url")
	pflag.StringVarP(&username, "username", "u", "TmdGj68KtaKDbCG5yTMmrB9mFKtzAQpBQqq.worker1", "username")
	pflag.StringVarP(&password, "password", "x", "x", "password")
	pflag.StringVarP(&loglevel, "loglevel", "l", "debug", "log level: info, debug, trace")
	pflag.StringVarP(&logfile, "logfile", "f", "debug.log", "logfile path")
	pflag.IntVarP(&threads, "threads", "t", runtime.NumCPU(), "threads")
	pflag.StringVarP(&algo, "algo", "a", "blake2bd", "algorithm: blake2bd")
	pflag.Parse()

	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
	if l, err := logrus.ParseLevel(loglevel); err != nil {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(l)
	}

	if _, ok := PowTypeMap[algo]; !ok {
		algo = "blake2bd"
	}
	logrus.Infof("algorithm set to %v", algo)

	if logfile == "" {
		logrus.Warningf("ignore logging to file")
	}
	ljack := &lumberjack.Logger{
		Filename: logfile,
	}
	mWriter := io.MultiWriter(os.Stdout, ljack)
	logrus.SetOutput(mWriter)

	powType := PowTypeMap[algo]
	switch powType {
	case BLAKE2BD:
		SetHashFunc(Blake2bD)
	case KECCAK256:
		SetHashFunc(QitmeerKeecak256)
	}

	cfg := &StratumMinerConfig{
		Url:      url,
		Username: username,
		Password: password,
		SumIntv:  "10s",
		Threads:  threads,
		PowType:  powType,
	}
	m := NewMiner(cfg)
	m.Mine()
}
