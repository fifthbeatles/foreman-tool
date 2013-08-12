package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

const APP_VERSION = "0.1"

var (
	versionFlag *bool = flag.Bool("v", false, "Print the version number.")

	mysqlHostFlag     *string = flag.String("mysql-host", "localhost", "Set MySQL host")
	mysqlPortFlag     *uint   = flag.Uint("mysql-port", 3306, "Set MySQL port")
	mysqlUserFlag     *string = flag.String("mysql-user", "foreman", "Set MySQL user")
	mysqlPasswordFlag *string = flag.String("mysql-password", "foreman", "Set MySQL passsword")
	mysqlDatabaseFlag *string = flag.String("mysql-database", "foreman", "Set MySQL database")

	hostsFileFlag *string = flag.String("f", "", "Set input file containing hosts(each line only has one host)")
	durationFlag  *string = flag.String("t", "1h", "Set duration between start and now like 15m, 1.5h, 2h45m(s for seconds, m for minutes, h for hours)")
)

func main() {
	flag.Parse()

	if *versionFlag {
		fmt.Println("Version:", APP_VERSION)
		os.Exit(0)
	}

	foreman := NewForeman(*mysqlHostFlag, *mysqlPortFlag, *mysqlUserFlag, *mysqlPasswordFlag, *mysqlDatabaseFlag)
	if err := foreman.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer foreman.Close()

	duration, err := time.ParseDuration(*durationFlag)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var hosts []string
	if len(*hostsFileFlag) > 0 {
		hosts, err = readHostsFile(*hostsFileFlag)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	foreman.Summary(duration, hosts)
}

func readHostsFile(filename string) (hosts []string, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	for {
		var line string
		line, err = reader.ReadString('\n')
		if err != nil {
			err = nil
			break
		}
		hosts = append(hosts, strings.Trim(line, "\r\n"))
	}
	return
}
