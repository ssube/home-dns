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

type Record struct {
	Cron string
	Name string
	TTL  int
	Zone string
}

type Config struct {
	Records []Record
	Source  string
}

func findPublicAddress(url string) []string {
	ipResp, ipError := http.Get(url)
	if ipError != nil {
		log.Printf("error getting public address: %s", ipError.Error())
	}
	defer ipResp.Body.Close()
	body, bodyError := ioutil.ReadAll(ipResp.Body)
	if bodyError != nil {
		log.Printf("error getting response body: %s", bodyError.Error())
	}

	return []string{
		string(body),
	}
}

func loadConfig(path string) (*Config, error) {
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

func mapResourceValues(values []string) []*route53.ResourceRecord {
	records := make([]*route53.ResourceRecord, len(values))
	for i, v := range values {
		records[i] = &route53.ResourceRecord{
			Value: aws.String(v),
		}
	}
	return records
}

func updateRecord(conf *Config, r Record) {
	// create a session
	sess, sessError := session.NewSession()
	if sessError != nil {
		log.Printf("error creating aws session: %s", sessError.Error())
		return
	}

	// get current ip
	values := findPublicAddress(conf.Source)
	log.Printf("updating '%s' to '%s'", r.Name, values)

	// prepare the update
	updates := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name:            aws.String(r.Name),
						TTL:             aws.Int64(300),
						Type:            aws.String("A"),
						ResourceRecords: mapResourceValues(values),
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
	}

	log.Printf("updated route53 records: %v", updateOutput)
}

func scheduleJob(conf *Config, c *cron.Cron, r Record) {
	log.Printf("scheduling cron job for %s", r.Name)
	c.AddFunc(r.Cron, func() {
		log.Printf("executing cron job for %s", r.Name)
		updateRecord(conf, r)
	})
}

func main() {
	if len(os.Args) < 2 {
		log.Printf("not enough arguments: ./home-dns config.yml")
		return
	}

	confPath := os.Args[1]
	log.Printf("loading config from '%s'", confPath)
	conf, confError := loadConfig(confPath)
	if confError != nil {
		log.Printf("error loading config: %s", confError.Error())
		return
	}

	// open a session (first, to fail early)

	// schedule cron jobs
	c := cron.New()
	for _, r := range conf.Records {
		scheduleJob(conf, c, r)
	}
	c.Start()

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt, os.Kill)
	<-stop

	c.Stop()
}
