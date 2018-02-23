package lib

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"strings"
	"math"
)

/*
	Parse platform file
*/

type topology struct {
	XMLName       xml.Name       `xml:"topology"`
	ID            string         `xml:"id,attr"`
	Storage_types []storage_type_xml `xml:"storage_type"`
	Hosts         []host_xml         `xml:"host"`
	Links         []link_xml         `xml:"link"`
	Routes        []route_xml        `xml:"route"`
	Storages      []storage_xml      `xml:"storage"`
}

type storage_type_xml struct {
	XMLName     xml.Name     `xml:"storage_type"`
	ID          string       `xml:"id,attr"`
	Size        string       `xml:"size,attr"`
	Model_props []model_prop_xml `xml:"model_prop"`
}

type model_prop_xml struct {
	XMLName xml.Name `xml:"model_prop"`
	Id      string   `xml:"id,attr"`
	Value   string   `xml:"value,attr"`
}

type storage_xml struct {
	XMLName xml.Name `xml:"storage"`
	Id      string   `xml:"id,attr"`
	TypeId  string   `xml:"typeId,attr"`
	Attach  string   `xml:"attach,attr"`
	Content string   `xml:"content,attr"`
}

type host_xml struct {
	XMLName xml.Name `xml:"host"`
	Id      string   `xml:"id,attr"`
	Speed   string   `xml:"speed,attr"`
	Mounts  []mount_xml  `xml:"mount"`
}

type mount_xml struct {
	XMLName   xml.Name `xml:"mount"`
	StorageId string   `xml:"storageId,attr"`
	Name      string   `xml:"name,attr"`
}

type link_xml struct {
	XMLName   xml.Name `xml:"link"`
	Id        string   `xml:"id,attr"`
	Bandwidth string   `xml:"bandwidth,attr"`
	Latency   string   `xml:"latency,attr"`
}

type route_xml struct {
	XMLName   xml.Name   `xml:"route"`
	Src       string     `xml:"src,attr"`
	Dst       string     `xml:"dst,attr"`
	Link_ctns []link_ctn_xml `xml:"link_ctn"`
}

type link_ctn_xml struct {
	XMLName xml.Name `xml:"link_ctn"`
	Id      string   `xml:"id,attr"`
}

func SIM_platform_init(FilePath string) {
	// Open our xmlFile
	xmlFile, err := os.Open(FilePath)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	var TOPOLOGY topology
	xml.Unmarshal(byteValue, &TOPOLOGY)

	platform := make(map[Route]*Link)
	hostsMap := make(map[string]*Host)
	storagesMap := make(map[string]*Storage)

	FunctionsMap := make(map[string]func(*Process, []string))

	// STORAGE_TYPES
	StorageTypes := make(map[string]*StorageType)
	for i := 0; i < len(TOPOLOGY.Storage_types); i++{
		s := StorageType{
			typeId:TOPOLOGY.Storage_types[i].ID,
			size:UnitToFloatParser(TOPOLOGY.Storage_types[i].Size),
		}
		for j := 0; j < len(TOPOLOGY.Storage_types[i].Model_props); j++{
			speed := TOPOLOGY.Storage_types[i].Model_props[j].Id
			value := TOPOLOGY.Storage_types[i].Model_props[j].Value
			if strings.Compare(speed, "read") == 0{
				s.readRate = UnitToFloatParser(value)
			}else if strings.Compare(speed, "write") == 0{
				s.writeRate = UnitToFloatParser(value)
			}
		}
		StorageTypes[TOPOLOGY.Storage_types[i].ID] = &s
	}

	// STORAGES
	for i := 0; i < len(TOPOLOGY.Storages); i++{
		name := TOPOLOGY.Storages[i].Id
		sType := TOPOLOGY.Storages[i].TypeId
		storage := &Storage{
			StorageType: StorageTypes[sType],
			name:name,
			usedSize:int64(0),
		}
		storagesMap[name] = storage
	}

	// HOSTS
	for i := 0; i < len(TOPOLOGY.Hosts); i++ {
		HostName := TOPOLOGY.Hosts[i].Id
		HostSpeed := UnitToFloatParser(TOPOLOGY.Hosts[i].Speed)
		Host := &Host{name: HostName,
			speed: HostSpeed,
		}

		//Mount storage to host
		// TODO
		for j := 0; j < len(TOPOLOGY.Hosts[i].Mounts); j++ {
			Host.storage = storagesMap[TOPOLOGY.Hosts[i].Mounts[j].StorageId]
		}
		hostsMap[HostName] = Host
	}

	env.hostsMap = hostsMap

	// ROUTES
	for i := 0; i < len(TOPOLOGY.Routes); i++ {
		SRCHost := env.getHostByName(TOPOLOGY.Routes[i].Src)
		DSTHost := env.getHostByName(TOPOLOGY.Routes[i].Dst)
		RealLinkId := TOPOLOGY.Routes[i].Link_ctns[0].Id

		RealLink := TOPOLOGY.getLinkById(RealLinkId)
		RealLinkBW := UnitToFloatParser(RealLink.Bandwidth)

		RealRoute := Route{
			start:  SRCHost,
			finish: DSTHost,
		}

		platform[RealRoute] = &Link{
			Resource: &Resource{
				bandwidth: RealLinkBW,
				mutex:     sync.Mutex{},
				queue:     []*TransferEvent{},
				counter:   0,
				env:       env,
			},
			name:RealLinkId,
		}
	}

	env.routesMap = platform
	env.FunctionsMap = FunctionsMap
	env.storagesMap = storagesMap

}

/*
	Parse deployment file
*/

type deployment struct {
	XMLName   xml.Name  `xml:"deployment"`
	Processes []process `xml:"process"`
}

type process struct {
	XMLName   xml.Name   `xml:"process"`
	Host      string     `xml:"host,attr"`
	Function  string     `xml:"function,attr"`
	Arguments []argument `xml:"argument"`
}

type argument struct {
	XMLName xml.Name `xml:"argument"`
	Value   string   `xml:"value,attr"`
}

func SIM_launch_application(FilePath string) {
	// Open our xmlFile
	xmlFile, err := os.Open(FilePath)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	var DEPLOYMENT deployment
	xml.Unmarshal(byteValue, &DEPLOYMENT)

	for i := 0; i < len(DEPLOYMENT.Processes); i++ {
		HostName := DEPLOYMENT.Processes[i].Host
		FuncName := DEPLOYMENT.Processes[i].Function
		Arguments := argsToStr(DEPLOYMENT.Processes[i].Arguments)
		Func := env.getFunctionByName(FuncName)
		Host := env.getHostByName(HostName)
		SIM_process_create_with_agrs(FuncName, Func, Host, nil, Arguments)
	}
}

func (TOPOLOGY *topology) getLinkById(ID string) link_xml {
	for i := 0; i < len(TOPOLOGY.Links); i++ {
		if TOPOLOGY.Links[i].Id == ID {
			return TOPOLOGY.Links[i]
		}
	}
	panic("No such link id")
	return link_xml{}
}

/*
 SIM function register
*/

func (env *Environment) getFunctionByName(FuncName string) func(*Process, []string) {
	Func, ok := env.FunctionsMap[FuncName]
	if !ok {
		panic(fmt.Sprintf("No such registered function %s", FuncName))
	}
	return Func
}

func argsToStr(args []argument) []string {
	array := make([]string, len(args))
	for i := range args {
		array[i] = args[i].Value
	}
	return array
}


func UnitToFloatParser(value string) float64{
	TERA := math.Pow10(12)
	GIGA := math.Pow10(9)
	MEGA := math.Pow10(6)
	KILO := math.Pow10(3)
	if strings.HasSuffix(value, "TB"){
		s := strings.TrimRight(value, "TB")
		converted, _ := strconv.ParseFloat(s, 64)
		return TERA * converted
	}else if strings.HasSuffix(value, "GB"){
		s := strings.TrimRight(value, "GB")
		converted, _ := strconv.ParseFloat(s, 64)
		return GIGA * converted
	} else if strings.HasSuffix(value, "MB"){
		s := strings.TrimRight(value, "MB")
		converted, _ := strconv.ParseFloat(s, 64)
		return MEGA * converted
	} else if strings.HasSuffix(value, "KB"){
		s := strings.TrimRight(value, "KB")
		converted, _ := strconv.ParseFloat(s, 64)
		return KILO * converted
	} else if strings.HasSuffix(value, "B"){
		s := strings.TrimRight(value, "B")
		converted, _ := strconv.ParseFloat(s, 64)
		return converted
	} else if strings.HasSuffix(value, "GBps"){
		s := strings.TrimRight(value, "GBps")
		converted, _ := strconv.ParseFloat(s, 64)
		return GIGA * converted
	} else if strings.HasSuffix(value, "MBps"){
		s := strings.TrimRight(value, "MBps")
		converted, _ := strconv.ParseFloat(s, 64)
		return MEGA * converted
	} else if strings.HasSuffix(value, "KBps"){
		s := strings.TrimRight(value, "KBps")
		converted, _ := strconv.ParseFloat(s, 64)
		return KILO * converted
	} else if strings.HasSuffix(value, "Bps"){
		s := strings.TrimRight(value, "Bps")
		converted, _ := strconv.ParseFloat(s, 64)
		return converted
	} else if strings.HasSuffix(value, "Gf"){
		s := strings.TrimRight(value, "Gf")
		converted, _ := strconv.ParseFloat(s, 64)
		return GIGA * converted
	} else if strings.HasSuffix(value, "Mf"){
		s := strings.TrimRight(value, "Mf")
		converted, _ := strconv.ParseFloat(s, 64)
		return MEGA * converted
	} else if strings.HasSuffix(value, "Kf"){
		s := strings.TrimRight(value, "Kf")
		converted, _ := strconv.ParseFloat(s, 64)
		return KILO * converted
	} else if strings.HasSuffix(value, "f"){
		s := strings.TrimRight(value, "f")
		converted, _ := strconv.ParseFloat(s, 64)
		return converted
	}
	panic("PARSED incorrectly")
}