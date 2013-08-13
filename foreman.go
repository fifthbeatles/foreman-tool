package main

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"sort"
	"time"
)

type Foreman struct {
	dsn   string
	db    *sql.DB
	hosts map[string]uint
}

func NewForeman(host string, port uint, user string, password string, database string) *Foreman {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", user, password, host, port, database)
	return &Foreman{dsn: dsn}
}

func (f *Foreman) Open() (err error) {
	if f.db, err = sql.Open("mysql", f.dsn); err != nil {
		return
	}
	if f.hosts == nil {
		f.hosts = make(map[string]uint)
	}
	rows, err := f.db.Query("SELECT name, id FROM hosts ORDER BY name")
	if err != nil {
		return
	}
	for rows.Next() {
		var id uint
		var name string
		err = rows.Scan(&name, &id)
		if err != nil {
			return
		}
		f.hosts[name] = id
	}
	return
}

func (f *Foreman) IsOpen() bool {
	return f.db != nil
}

func (f *Foreman) Close() (err error) {
	return f.db.Close()
}

func (f *Foreman) hostId(name string) (id uint, err error) {
	if f.hosts == nil {
		err = errors.New("hosts map nil")
		return
	}
	id, ok := f.hosts[name]
	if !ok {
		err = errors.New("non-exist host " + name)
		return
	}
	return
}

func (f *Foreman) Summary(duration time.Duration, hosts []string) {
	reports, err := f.reports(duration)
	if err != nil {
		fmt.Println(err)
		return
	}
	var missing_hosts, ok_hosts, error_hosts []string
	if hosts == nil || len(hosts) == 0 {
		for host, _ := range f.hosts {
			host_reports, ok := reports[host]
			if !ok {
				missing_hosts = append(missing_hosts, host)
				continue
			}
			ok = false
			for _, report := range host_reports {
				if report[1] < 4096 {
					ok_hosts = append(ok_hosts, host)
					ok = true
					break
				}
			}
			if !ok {
				error_hosts = append(error_hosts, host)
			}
		}
	} else {
		for _, host := range hosts {
			host_reports, ok := reports[host]
			if !ok {
				missing_hosts = append(missing_hosts, host)
				continue
			}
			ok = false
			for _, report := range host_reports {
				if report[1] < 4096 {
					ok_hosts = append(ok_hosts, host)
					ok = true
					break
				}
			}
			if !ok {
				error_hosts = append(error_hosts, host)
			}
		}
	}
	fmt.Printf("Missing Hosts(%d):\n", len(missing_hosts))
	sort.Strings(missing_hosts)
	for _, host := range missing_hosts {
		fmt.Println(host)
	}
	fmt.Printf("Error Hosts(%d):\n", len(error_hosts))
	sort.Strings(error_hosts)
	for _, host := range error_hosts {
		fmt.Println(host)
	}
	fmt.Printf("Good Hosts(%d):\n", len(ok_hosts))
	sort.Strings(ok_hosts)
	for _, host := range ok_hosts {
		fmt.Println(host)
	}
}

func (f *Foreman) reports(duration time.Duration) (reports map[string][][2]int, err error) {
	startTime := time.Now().UTC().Add(-1 * duration).Format("2006-01-02 15:04:05")
	sql := fmt.Sprintf("select hosts.name, reports.id, reports.status from reports inner join hosts on (hosts.id = reports.host_id) where reports.reported_at > '%s' order by reports.id desc;", startTime)
	rows, err := f.db.Query(sql)
	if err != nil {
		return
	}
	reports = make(map[string][][2]int)
	for rows.Next() {
		var host string
		var status int
		var id int
		err = rows.Scan(&host, &id, &status)
		if err != nil {
			return
		}
		reports[host] = append(reports[host], [2]int{id, status})
	}
	return
}
