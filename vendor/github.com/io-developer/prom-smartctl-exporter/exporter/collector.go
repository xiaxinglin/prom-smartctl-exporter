package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"os"
	"os/exec"
)

var _ prometheus.Collector = &Collector{}

type Collector struct {
	device       string
	PowerOnHours *prometheus.Desc
	Reallocated_Sector_Ct *prometheus.Desc
	Reported_Uncorrect *prometheus.Desc
	Command_Timeout *prometheus.Desc
	Load_Cycle_Count *prometheus.Desc
	Temperature  *prometheus.Desc
	Current_Pending_Sector *prometheus.Desc
	Offline_Uncorrectable *prometheus.Desc 
	Total_LBAs_Written *prometheus.Desc 
	Total_LBAs_Read *prometheus.Desc 
}

func NewCollector(device string) *Collector {
	var (
		labels = []string{
			"device",
			"serial",
			"model",
			"host",
		}
	)
	return &Collector{
		device: device,
		PowerOnHours: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "power_on_hours"),
			"Power on hours",
			labels,
			nil,
		),
		Reallocated_Sector_Ct: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "Reallocated_Sector_Ct"),
			"Reallocated_Sector_Ct",
			labels,
			nil,
		),
		Reported_Uncorrect: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "Reported_Uncorrect"),
			"Reported_Uncorrect",
			labels,
			nil,
		),
		Command_Timeout: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "Command_Timeout"),
			"Command_Timeout",
			labels,
			nil,
		),
		Load_Cycle_Count: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "Load_Cycle_Count"),
			"Load_Cycle_Count",
			labels,
			nil,
		),
		Temperature: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "temperature"),
			"Temperature",
			labels,
			nil,
		),
		Current_Pending_Sector: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "Current_Pending_Sector"),
			"Current_Pending_Sector",
			labels,
			nil,
		),
		Offline_Uncorrectable: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "Offline_Uncorrectable"),
			"Offline_Uncorrectable",
			labels,
			nil,
		),
		Total_LBAs_Written: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "Total_LBAs_Written"),
			"Total_LBAs_Written",
			labels,
			nil,
		),
		Total_LBAs_Read: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "Total_LBAs_Read"),
			"Total_LBAs_Read",
			labels,
			nil,
		),
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ds := []*prometheus.Desc{
		c.PowerOnHours,
		c.Reallocated_Sector_Ct,
		c.Reported_Uncorrect,
		c.Command_Timeout,
		c.Load_Cycle_Count,
		c.Temperature,
		c.Current_Pending_Sector,
		c.Offline_Uncorrectable,
		c.Total_LBAs_Written,
		c.Total_LBAs_Read,
	}
	for _, d := range ds {
		ch <- d
	}
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	if desc, err := c.collect(ch); err != nil {
		log.Printf("[ERROR] failed collecting metric %v: %v", desc, err)
		ch <- prometheus.NewInvalidMetric(desc, err)
		return
	}
}

func (c *Collector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	if c.device == "" {
		return nil, nil
	}
	
	out, err := exec.Command("smartctl", "-iA", c.device).CombinedOutput()
	if err != nil {
		log.Printf("[ERROR] smart log: \n%s\n", out)
		return nil, err
	}
	smart := ParseSmart(string(out))

	serialNumber := smart.GetInfo("Serial Number")
	if serialNumber == ""{
		return nil, nil
	}
	hostName, _ := os.Hostname()

	labels := []string{
		c.device,
		serialNumber,
		smart.GetInfo("Device Model"),
		hostName,
	}

	ch <- prometheus.MustNewConstMetric(
		c.PowerOnHours,
		prometheus.GaugeValue,
		float64(smart.GetAttr(9).rawValue),
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.Reallocated_Sector_Ct,
		prometheus.GaugeValue,
		float64(smart.GetAttr(5).rawValue),
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.Reported_Uncorrect,
		prometheus.GaugeValue,
		float64(smart.GetAttr(187).rawValue),
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.Command_Timeout,
		prometheus.GaugeValue,
		float64(smart.GetAttr(188).rawValue),
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.Load_Cycle_Count,
		prometheus.GaugeValue,
		float64(smart.GetAttr(193).rawValue),
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.Temperature,
		prometheus.GaugeValue,
		float64(smart.GetAttr(194).rawValue),
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.Current_Pending_Sector,
		prometheus.GaugeValue,
		float64(smart.GetAttr(197).rawValue),
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.Offline_Uncorrectable,
		prometheus.GaugeValue,
		float64(smart.GetAttr(198).rawValue),
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.Total_LBAs_Written,
		prometheus.GaugeValue,
		float64(smart.GetAttr(241).rawValue),
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.Total_LBAs_Read,
		prometheus.GaugeValue,
		float64(smart.GetAttr(242).rawValue),
		labels...,
	)

	return nil, nil
}
