package main

import (
	"GodDns/DDNS"
	"GodDns/Device"
	log "GodDns/Log"
	"GodDns/Net"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli/v2"

	_ "GodDns/Service" // register all services
)

var output = os.Stdout

const MAXRETRY = 255
const defaultRetryAttempt = 3
const MINTIMEGAP = 5

const (
	run             = "run"
	runAuto         = "run-auto"
	runApi          = "run-api"
	runAutoOverride = "run-auto-override"
)

var (
	Time              uint64 = 0
	TimeLimitation    uint64 = 0 // 0 means no limitation
	ApiName                  = ""
	retryAttempt      uint8  = defaultRetryAttempt
	config                   = ""
	defaultLocation          = ""
	logLevel                 = "Info"
	proxy                    = ""
	proxyEnable              = false
	parallelExecuting        = false
	// cleanUp         func()
	runMode = ""
)

var (
	silentFlag = &cli.BoolFlag{
		Name:    "silent",
		Aliases: []string{"s", "S"},
		Value:   false,
		Usage:   "no message output",
		Action: func(context *cli.Context, silent bool) error {
			// set output
			if silent {
				output = nil
			}
			return nil
		},
	}

	timeFlag = &cli.Uint64Flag{
		Name:        "time",
		Aliases:     []string{"t", "T"},
		Value:       0,
		Usage:       "run ddns per time(`seconds`)",
		Destination: &Time,
		Action: func(context *cli.Context, u uint64) error {
			if u < MINTIMEGAP {
				return errors.New("time gap is too short")
			}
			return nil
		},
	}

	timeLimitationFlag = &cli.Uint64Flag{
		Name:        "time-limitation",
		Aliases:     []string{"tl", "TL"},
		Value:       0,
		Usage:       "run ddns per time(seconds) up to n `times`",
		Destination: &TimeLimitation,
		Action: func(context *cli.Context, u uint64) error {
			t := context.Uint64("time")
			if t == 0 {
				return errors.New("time limitation must be used with time flag")
			}
			return nil
		},
	}

	retryFlag = &cli.UintFlag{
		Name:  "retry",
		Value: defaultRetryAttempt,
		Usage: "retry `times`",
		Action: func(context *cli.Context, u uint) error {
			if u > MAXRETRY {
				return fmt.Errorf("too many retry times, should be less than %d", MAXRETRY)
			}
			retryAttempt = uint8(u)
			return nil
		},
	}

	logFlag = &cli.StringFlag{
		Name:        "log",
		Aliases:     []string{"l", "L", "Log"},
		Value:       "Info",
		Usage:       "`level`: Trace/Debug/Info/Warn/Error",
		Destination: &logLevel,
	}

	configFlag = &cli.StringFlag{
		Name:        "config",
		Aliases:     []string{"c", "C", "Config"},
		Value:       "",
		DefaultText: defaultLocation,
		Usage:       "set configuration `file`",
		Destination: &config,
	}

	proxyFlag = &cli.StringFlag{
		Name:        "proxy",
		Aliases:     []string{"p", "P", "Proxy"},
		Value:       "",
		Usage:       "set proxy `url`",
		Destination: &proxy,
		Action: func(context *cli.Context, s string) error {
			if s != "" {
				if s == "enable" {
					proxyEnable = true
					return nil
				} else if s == "disable" {
					return nil
				} else if Net.IsProxyValid(s) {
					proxyEnable = true
					Net.AddProxy2Top(Net.GlobalProxys, s)
					return nil
				} else {
					return errors.New("invalid proxy url")
				}
			}
			return errors.New("empty proxy url")
		},
	}

	parallelFlag = &cli.BoolFlag{
		Name:        "parallel",
		Aliases:     []string{"Parallel"},
		Value:       false,
		Usage:       "run ddns parallel",
		Destination: &parallelExecuting,
	}
)

func init() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"v", "V"},
		Usage:   "print the version info/upgrade info",
	}

	cli.VersionPrinter = func(c *cli.Context) {
		msg := make(chan string, 2)
		go CheckVersionUpgrade(msg)
		fmt.Println(DDNS.NowVersionInfo())
		for i := 0; i < 2; i++ {
			select {
			case s := <-msg:
				if s != "" {
					fmt.Println(s)
				}
			case <-time.After(2 * time.Second):
				return
			}
		}

	}

	cli.HelpFlag = &cli.BoolFlag{
		Name:    "help",
		Aliases: []string{"h", "H"},
		Usage:   "show help",
	}

	var err error
	defaultLocation, err = DDNS.GetDefaultConfigurationLocation()
	if err != nil {
		defaultLocation = "./DDNS.conf"
	}

}

func checkLog(l string) error {
	switch l {
	case "Trace", "trace", "TRACE":
		fallthrough
	case "Debug", "debug", "DEBUG":
		fallthrough
	case "Info", "info", "INFO":
		fallthrough
	case "Warn", "warn", "WARN":
		fallthrough
	case "Error", "error", "ERROR":
		_, err := log.InitLog("DDNS.log", 0666, l, output)
		if err != nil {
			log.Error("failed to init log file ", log.String("error", err.Error()))
			return err
		}
		// cleanUp = clean
		return nil
	default:
		return errors.New("invalid log level")
	}
}

// todo return config setting command `GodDns config -service=cloudflare`
func main() {

	var parameters []DDNS.Parameters
	var GlobalDevice Device.Device
	configFactoryList := DDNS.ConfigFactoryList

	location, err := DDNS.GetProgramConfigLocation()
	if err != nil {
		_, _ = fmt.Fprintln(output, "error loading program config: ", err, " use default config")
	} else {
		if DDNS.IsConfigExist(location) {
			programConfig, fatal, other := DDNS.LoadProgramConfig(location)
			if fatal != nil {
				// default setup
				_, _ = fmt.Fprintln(output, "error loading program config: ", err, " use default config")
				_, _ = fmt.Fprintln(output, fatal)
				DDNS.DefaultConfig.Setup()
			} else {
				if other != nil {
					_, _ = fmt.Fprintln(output, other)
				}
				programConfig.Setup()
			}
		} else {
			// create Config here
			_, _ = fmt.Fprintln(output, "no config at ", location, " try to generate a default config")
			err := DDNS.DefaultConfig.GenerateConfigFile()
			DDNS.DefaultConfig.Setup()
			if err != nil {
				_, _ = fmt.Fprintln(output, "failed to generate default program config at ", location)
			} else {
				_, _ = fmt.Fprintln(output, "generate default program config at ", location)
			}
		}
	}

	app := &cli.App{
		Name:     DDNS.FullName,
		Usage:    "A DDNS tool written in Go",
		Version:  DDNS.NowVersion.Info(),
		Compiled: time.Now(),
		Authors: []*cli.Author{
			{
				Name:  DDNS.Author,
				Email: DDNS.FeedbackEmail(),
			},
		},
		Suggest:              true,
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:    "run",
				Aliases: []string{"r", "R"},
				Usage:   "run the DDNS service",

				Action: func(context *cli.Context) error {
					err := checkLog(logLevel)
					if err != nil {
						return err
					}

					if config != "" {
						DDNS.UpdateConfigureLocation(config)
					} else {
						DDNS.UpdateConfigureLocation(defaultLocation)
					}

					parametersTemp, err := ReadConfig(configFactoryList)
					if err != nil {
						return err
					}
					parameters = parametersTemp

					if ApiName == "" {
						runMode = run
					} else {
						runMode = runApi
					}

					if Time != 0 {
						_ = RunDDNS(parameters)
						RunPerTime(Time, nil, parameters)
					}

					return RunDDNS(parameters)
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "api",
						Aliases: []string{"i", "I"},

						Usage: "get ip address from provided `ApiName`, eg: ipify/identMe",

						Destination: &ApiName,
					},
					parallelFlag,
					timeFlag,
					timeLimitationFlag,
					retryFlag,
					silentFlag,
					logFlag,
					configFlag,
					proxyFlag,
				},
				Subcommands: []*cli.Command{
					{
						Name:    "auto",
						Aliases: []string{"a", "A"},
						Usage:   "run ddns, use ip address of interface set in Device Section automatically",
						Action: func(context *cli.Context) error {

							err := checkLog(logLevel)
							if err != nil {
								return err
							}

							if config != "" {
								DDNS.UpdateConfigureLocation(config)
							} else {
								DDNS.UpdateConfigureLocation(defaultLocation)
							}

							parametersTemp, err := ReadConfig(configFactoryList)
							if err != nil {
								return err
							}
							parameters = parametersTemp
							GlobalDevice, err = GetGlobalDevice(parameters)
							if err != nil {
								return err
							}

							runMode = runAuto

							if Time != 0 {
								_ = RunAuto(GlobalDevice, parameters)
								RunPerTime(Time, &GlobalDevice, parameters)
							}

							return RunAuto(GlobalDevice, parameters)
						},
						Flags: []cli.Flag{
							parallelFlag,
							timeFlag,
							timeLimitationFlag,
							retryFlag,
							silentFlag,
							logFlag,
							configFlag,
							proxyFlag,
						},
						Subcommands: []*cli.Command{
							{
								Name:    "override",
								Aliases: []string{"o", "O"},
								Usage:   "run ddns, override the ip address of interface set in each service Section",
								Flags: []cli.Flag{
									parallelFlag,
									timeFlag,
									timeLimitationFlag,
									retryFlag,
									silentFlag,
									logFlag,
									configFlag,
									proxyFlag,
								},
								Action: func(context *cli.Context) error {

									err := checkLog(logLevel)
									if err != nil {
										return err
									}

									if config != "" {
										DDNS.UpdateConfigureLocation(config)
									} else {
										DDNS.UpdateConfigureLocation(defaultLocation)
									}

									parametersTemp, err := ReadConfig(configFactoryList)
									if err != nil {
										return err
									}
									parameters = parametersTemp
									GlobalDevice, err = GetGlobalDevice(parameters)
									if err != nil {
										return err
									}

									runMode = runAutoOverride

									if Time != 0 {
										_ = RunOverride(GlobalDevice, parameters)
										RunPerTime(Time, &GlobalDevice, parameters)
									}

									return RunOverride(GlobalDevice, parameters)
								},
							},
						},
					},
				},
			},
			{

				Name:    "generate",
				Aliases: []string{"g", "G"},
				Usage:   "generate a default configuration file",
				Action: func(*cli.Context) error {
					err := checkLog(logLevel)
					if err != nil {
						return err
					}

					if config != "" {
						DDNS.UpdateConfigureLocation(config)
					} else {
						DDNS.UpdateConfigureLocation(defaultLocation)
					}
					return GenerateConfigure(configFactoryList)
				},
				Flags: []cli.Flag{
					silentFlag,
					logFlag,
					configFlag,
				},
			},
		},
		After: func(context *cli.Context) error {
			// if cleanUp != nil {
			// 	// bug cleanUp()
			// }
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Errorf("fatal: %s", err)
	}

}
