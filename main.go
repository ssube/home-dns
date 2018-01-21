package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"

	"gopkg.in/robfig/cron.v2"
	yaml "gopkg.in/yaml.v2"
)

// Record defines a single record to upsert and the schedule for updates
type Record struct {
	Cron string
	Name string
	TTL  int64
	Zone string
}

// Config defines the source endpoint and a set of records to update
type Config struct {
	Records []Record
	Source  string
}

// FindPublicAddress fetches the public address from an external HTTP/S endpoint
func FindPublicAddress(url string) ([]string, error) {
	ipResp, ipError := http.Get(url)
	if ipError != nil {
		log.Printf("error getting public address: %s", ipError.Error())
		return nil, ipError
	}

	defer ipResp.Body.Close()
	body, bodyError := ioutil.ReadAll(ipResp.Body)
	if bodyError != nil {
		log.Printf("error getting response body: %s", bodyError.Error())
		return nil, bodyError
	}

	return []string{
		string(body),
	}, nil
}

// LoadConfig loads the config data (source and records) from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, readError := ioutil.ReadFile(path)
	if readError != nil {

	}

	dest := &Config{}
	yamlError := yaml.Unmarshal(data, dest)
	if yamlError != nil {
		log.Printf("error parsing config: %s", yamlError.Error())
		return nil, yamlError
	}

	return dest, nil
}

// MapResourceValues converts address strings to route53's special record type
func MapResourceValues(values []string) []*route53.ResourceRecord {
	records := make([]*route53.ResourceRecord, len(values))
	for i, v := range values {
		records[i] = &route53.ResourceRecord{
			Value: aws.String(v),
		}
	}
	return records
}

// UpdateRecord upserts a single record in a zone
func UpdateRecord(conf *Config, r Record) {
	// create a session
	sess, sessError := session.NewSession()
	if sessError != nil {
		log.Printf("error creating aws session: %s", sessError.Error())
		return
	}

	// get current ip
	values, valueError := FindPublicAddress(conf.Source)
	if valueError != nil {
		log.Printf("error fetching external address: %s", valueError.Error())
		return
	}

	log.Printf("updating '%s' to '%s'", r.Name, values)

	// prepare the update
	updates := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name:            aws.String(r.Name),
						TTL:             aws.Int64(r.TTL),
						Type:            aws.String("A"),
						ResourceRecords: MapResourceValues(values),
					},
				},
			},
		},
		HostedZoneId: aws.String(r.Zone),
	}

	// apply
	svc := route53.New(sess)
	updateOutput, updateError := svc.ChangeResourceRecordSets(updates)
	if updateError != nil {
		log.Printf("error updating route53 records: %s", updateError.Error())
		return
	}

	log.Printf("updated route53 records: %v", updateOutput)
}

// ScheduleJob adds a new record and update job to the cron pool
func ScheduleJob(conf *Config, c *cron.Cron, r Record) {
	log.Printf("scheduling cron job for %s", r.Name)
	c.AddFunc(r.Cron, func() {
		log.Printf("executing cron job for %s", r.Name)
		UpdateRecord(conf, r)
	})
}

func main() {
	if len(os.Args) < 2 {
		log.Printf("not enough arguments: ./home-dns config.yml")
		return
	}

	confPath := os.Args[1]
	log.Printf("loading config from '%s'", confPath)
	conf, confError := LoadConfig(confPath)
	if confError != nil {
		log.Printf("error loading config: %s", confError.Error())
		return
	}

	// open a session (first, to fail early)

	// schedule cron jobs
	c := cron.New()
	for _, r := range conf.Records {
		ScheduleJob(conf, c, r)
	}
	c.Start()

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt, os.Kill)
	<-stop

	c.Stop()
}
