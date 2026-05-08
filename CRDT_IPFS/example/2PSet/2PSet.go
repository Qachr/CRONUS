package Set

import (
	CRDTDag "IPFS_CRDT/CRDTDag"
	"IPFS_CRDT/Config"
	CRDT "IPFS_CRDT/Crdt"
	Payload "IPFS_CRDT/Payload"
	IpfsLink "IPFS_CRDT/ipfsLink"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"golang.org/x/sync/semaphore"
)

// =======================================================================================
// Payload - OpBased
// =======================================================================================

type Element string
type OpNature int

const (
	ADD OpNature = iota
	REMOVE
)

type Operation struct {
	Elem Element
	Op   OpNature
}

func (self Operation) ToString() string {
	b, err := json.Marshal(self)
	if err != nil {
		panic(fmt.Errorf("Set Operation To string fail to Marshal\nError: %s", err))
	}
	return string(b[:])
}
func (op *Operation) op_from_string(s string) {
	err := json.Unmarshal([]byte(s), op)
	if err != nil {
		panic(fmt.Errorf("Set Operation To string fail to Marshal\nError: %s", err))
	}
}

type PayloadOpBased struct {
	Op Operation
	Id string
}

func (self *PayloadOpBased) Create_PayloadOpBased(s string, o1 Operation) {

	self.Op = o1
	self.Id = s
}
func (self *PayloadOpBased) ToString() string {
	b, err := json.Marshal(self)
	if err != nil {
		panic(fmt.Errorf("Set Operation To string fail to Marshal\nError: %s", err))
	}
	return string(b[:])
}
func (self *PayloadOpBased) FromString(s string) {
	err := json.Unmarshal([]byte(s), self)
	if err != nil {
		panic(fmt.Errorf("Set Operation To string fail to Marshal\nError: %s", err))
	}
}

// =======================================================================================
// CRDTSet OpBased
// =======================================================================================

type CRDTSetOpBased struct {
	sys     *IpfsLink.IpfsLink
	added   []string
	removed []string
}

func Create_CRDTSetOpBased(s *IpfsLink.IpfsLink) CRDTSetOpBased {
	return CRDTSetOpBased{
		sys:     s,
		added:   make([]string, 0),
		removed: make([]string, 0),
	}
}
func search(list []string, x string) int {
	for i := 0; i < len(list); i++ {
		if list[i] == x {
			return i
		}
	}
	return -1
}
func (self *CRDTSetOpBased) Add(x string) {
	if search(self.added, x) == -1 {
		self.added = append(self.added, x)
	}
}

func (self *CRDTSetOpBased) Remove(x string) {
	if search(self.removed, x) == -1 {
		self.removed = append(self.removed, x)
	}
}

func (self *CRDTSetOpBased) Lookup() []string {
	i := make([]string, 0)
	fmt.Println("size", len(self.added))
	for x := range self.added {
		if search(self.removed, self.added[x]) == -1 {
			i = append(i, self.added[x])
			i = append(i, ",")
		}
	}

	return i
}

func (self *CRDTSetOpBased) ToFile(file string) {

	b, err := json.Marshal(self)
	if err != nil {
		panic(fmt.Errorf("CRDTDagNode - ToFile Could not Marshall %s\nError: %s", file, err))
	}
	f, err := os.Create(file)
	if err != nil {
		panic(fmt.Errorf("CRDTDagNode - ToFile Could not Create the file %s\nError: %s", file, err))
	}
	f.Write(b)
	err = f.Close()
	if err != nil {
		panic(fmt.Errorf("CRDTDagNode - ToFile Could not Write to the file %s\nError: %s", file, err))
	}
}

// =======================================================================================
// CRDTSetDagNode OpBased
// =======================================================================================

type CRDTSetOpBasedDagNode struct {
	DagNode CRDTDag.CRDTDagNode
}

func (self *CRDTSetOpBasedDagNode) FromFile(fil string) {
	var pl Payload.Payload = &PayloadOpBased{}
	self.DagNode.CreateNodeFromFile(fil, &pl)
}

func (self *CRDTSetOpBasedDagNode) GetDirect_dependency() []CRDTDag.EncodedStr {

	return self.DagNode.DirectDependency
}

func (self *CRDTSetOpBasedDagNode) ToFile(file string) {

	self.DagNode.ToFile(file)
}
func (self *CRDTSetOpBasedDagNode) GetEvent() *Payload.Payload {

	return self.DagNode.Event
}
func (self *CRDTSetOpBasedDagNode) GetPiD() string {

	return self.DagNode.PID
}
func (self *CRDTSetOpBasedDagNode) CreateEmptyNode() *CRDTDag.CRDTDagNodeInterface {
	n := CreateDagNode(Operation{}, "")
	var node CRDTDag.CRDTDagNodeInterface = &n
	return &node
}
func CreateDagNode(o Operation, id string) CRDTSetOpBasedDagNode {
	var pl Payload.Payload = &PayloadOpBased{Op: o, Id: id}
	slic := make([]CRDTDag.EncodedStr, 0)
	return CRDTSetOpBasedDagNode{
		DagNode: CRDTDag.CRDTDagNode{
			Event:            &pl,
			PID:              id,
			DirectDependency: slic,
		},
	}
}

// =======================================================================================
// CRDTSetDag OpBased
// =======================================================================================

type CRDTSetOpBasedDag struct {
	dag         CRDTDag.CRDTManager
	measurement bool
}

func (self *CRDTSetOpBasedDag) GetDag() *CRDTDag.CRDTManager {

	return &self.dag
}
func (self *CRDTSetOpBasedDag) SendRemoteUpdates() {

	self.dag.SendRemoteUpdates()
}
func (self *CRDTSetOpBasedDag) GetCRDTManager() *CRDTDag.CRDTManager {

	return &self.dag
}
func (self *CRDTSetOpBasedDag) IsKnown(cid CRDTDag.EncodedStr) bool {

	find := false
	for x := range self.dag.GetAllNodes() {
		if string(self.dag.GetAllNodes()[x]) == string(cid.Str) {
			find = true
			break
		}
	}
	return find
}

// CRDTDag.TimingMeasure
// func (self *CRDTSetOpBasedDag) Merge(cids []CRDTDag.EncodedStr) ([]string, []([]byte), CRDTDag.TimingMeasure) {
func (self *CRDTSetOpBasedDag) Merge(cids []CRDTDag.EncodedStr) ([]string, []([]byte)) {
	times := CRDTDag.TimingMeasure{0, 0, 0, 0, 0, 0, CRDTDag.TimingMeasure2{0, 0, 0, 0}}
	timeReadCIDs := time.Now()
	to_add := make([]CRDTDag.EncodedStr, 0)
	for _, cid := range cids {
		find := self.IsKnown(cid)
		if !find {
			to_add = append(to_add, cid)
		}
	}
	times.TimeReadinMerge = (int(time.Since(timeReadCIDs).Nanoseconds()))

	Time_GetinMerge := time.Now()
	fils, err := self.dag.GetNodeFromEncodedCid(to_add)
	if err != nil {
		panic(fmt.Errorf("could not get ndoes from encoded cids\nerror :%s", err))
	}
	times.TimeGetinMerge = (int(time.Since(Time_GetinMerge).Nanoseconds()))

	Time_remoteaddMerge := time.Now()

	receivedFiles := make([]string, 0)
	receivedCids := make([]([]byte), 0)

	for i, f := range fils {
		receivedFiles = append(receivedFiles, f)
		receivedCids = append(receivedCids, to_add[i].Str)
	}
	times.TimeCreateDAGNODE = 0
	times.TimeFromFile = 0
	times.TimeremoteAddNodefor = 0
	for index := range fils {
		timeMeasured := time.Now()
		fil := fils[index]
		n := CreateDagNode(Operation{}, "") // Create an Empty operation

		//Measure 1
		times.TimeCreateDAGNODE = times.TimeCreateDAGNODE + (int(time.Since(timeMeasured).Nanoseconds()))
		timeMeasured = time.Now()

		n.FromFile(fil) // Fill it with the operation just read

		//Measure 2
		times.TimeFromFile = times.TimeFromFile + (int(time.Since(timeMeasured).Nanoseconds()))
		timeMeasured = time.Now()

		// addi_file, addi_CID, times2 := self.remoteAddNode(cids[index], n) // Add the data as a Remote operation (which are applied as a local one)

		addi_file, addi_CID := self.remoteAddNode(cids[index], n)
		times2 := CRDTDag.TimingMeasure2{}

		times.Timings.CheckDependency = times.Timings.CheckDependency + times2.CheckDependency

		times.Timings.CreateNodeFromFile = times.Timings.CreateNodeFromFile + times2.CreateNodeFromFile

		times.Timings.GetNodeFromEncoded = times.Timings.GetNodeFromEncoded + times2.GetNodeFromEncoded

		times.Timings.TimeAddNodeInCRDTDAG = times.Timings.TimeAddNodeInCRDTDAG + times2.TimeAddNodeInCRDTDAG

		for i := range addi_file {
			receivedFiles = append(receivedFiles, addi_file[i])
			receivedCids = append(receivedCids, addi_CID[i])
		}

		//Measure 3
		times.TimeremoteAddNodefor = times.TimeremoteAddNodefor + (int(time.Since(timeMeasured).Nanoseconds()))
		timeMeasured = time.Now()

	}
	times.ForloopinMerge = (int(time.Since(Time_remoteaddMerge).Nanoseconds()))
	// return receivedFiles, receivedCids, times
	return receivedFiles, receivedCids
}

// Return the additionnal files that have been added as dependency of the CID cID in newnode, if they have been added
// func (self *CRDTSetOpBasedDag) remoteAddNode(cID CRDTDag.EncodedStr, newnode CRDTSetOpBasedDagNode) ([]string, []([]byte), CRDTDag.TimingMeasure2) {
func (self *CRDTSetOpBasedDag) remoteAddNode(cID CRDTDag.EncodedStr, newnode CRDTSetOpBasedDagNode) ([]string, []([]byte)) {
	var pl CRDTDag.CRDTDagNodeInterface = &newnode
	// foundFILES, additionnalCIDs, times2 := self.dag.RemoteAddNodeSuper(cID, &pl)
	foundFILES, additionnalCIDs := self.dag.RemoteAddNodeSuper(cID, &pl)
	// return foundFILES, additionnalCIDs, times2
	return foundFILES, additionnalCIDs
}

func (thisCRDTDag *CRDTSetOpBasedDag) callAddToIPFS(bytes []byte, file string) (blocks.Block, error) {
	time_toencrypt := -1
	ti := time.Now()
	var path blocks.Block
	var err error
	if thisCRDTDag.dag.Key != "" {
		path, err = thisCRDTDag.GetCRDTManager().AddToIPFS(thisCRDTDag.dag.Sys, bytes, &time_toencrypt)
	} else {
		path, err = thisCRDTDag.GetCRDTManager().AddToIPFS(thisCRDTDag.dag.Sys, bytes)
		time_toencrypt = 0
	}
	if err != nil {
		panic(fmt.Errorf("error in callAddToIPFS, Couldn't add file to IPFS\nError: %s\n \t", err))
	}
	Total_AddTime := int(time.Since(ti).Nanoseconds())
	time_add := Total_AddTime - time_toencrypt

	if thisCRDTDag.measurement {
		// Write time to encrypt in a file
		fstrBis := ""
		if thisCRDTDag.dag.Key != "" {
			fstrBis = file + ".timeEncrypt"
			if _, err := os.Stat(fstrBis); !errors.Is(err, os.ErrNotExist) {
				os.Remove(fstrBis)
			}
			fil, err := os.OpenFile(fstrBis, os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				panic(fmt.Errorf("error RemoteAddNodeSupde - , Could not open the time file to write encoded data\nError: %s", err))
			}
			_, err = fil.Write([]byte(strconv.Itoa(time_toencrypt)))
			if err != nil {
				panic(fmt.Errorf("error RemoteAddNodeSupde - , Could not write the time file to write encoded data\nError: %s", err))
			}
			err = fil.Close()
			if err != nil {
				panic(fmt.Errorf("error RemoteAddNodeSupde - , Could not close the time file to write encoded data \nError: %s", err))
			}
		}

		// Write time to add to IFPS
		fstrBis = file + ".timeAdd"
		if _, err := os.Stat(fstrBis); !errors.Is(err, os.ErrNotExist) {
			os.Remove(fstrBis)
		}
		fil, err := os.OpenFile(fstrBis, os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			panic(fmt.Errorf("error RemoteAddNodeSupde - , Could not open the time file to write encoded data\nError: %s", err))
		}
		_, err = fil.Write([]byte(strconv.Itoa(time_add)))
		if err != nil {
			panic(fmt.Errorf("error RemoteAddNodeSupde - , Could not write the time file to write encoded data\nError: %s", err))
		}
		err = fil.Close()
		if err != nil {
			panic(fmt.Errorf("error remoteAddNodeSupde - , Could not close the time file to write encoded data \nError: %s", err))
		}
	}

	return path, err
}

func (self *CRDTSetOpBasedDag) Add(x string) (string, TimeTuple) {
	newNode := CreateDagNode(Operation{Elem: Element(x), Op: ADD}, self.GetSys().IpfsNode.Identity.ShortString())
	for dependency := range self.dag.Root_nodes {
		// fmt.Println("dep:", self.dag.Root_nodes[dependency].Str)
		newNode.DagNode.DirectDependency = append(newNode.DagNode.DirectDependency, self.dag.Root_nodes[dependency])
	}

	strFile := self.dag.NextFileName()
	if _, err := os.Stat(strFile); !errors.Is(err, os.ErrNotExist) {
		os.Remove(strFile)
	}
	newNode.ToFile(strFile)
	bytes, err := os.ReadFile(strFile)
	if err != nil {
		panic(fmt.Errorf("ERROR INCREMENT CRDTSetOpBasedDag, could not read file\nerror: %s", err))
	}
	path, err := self.callAddToIPFS(bytes, strFile)
	if err != nil {
		panic(fmt.Errorf("CRDTSetOpBasedDag Increment, could not add the file to IFPS\nerror: %s", err))
	}

	encodedCid := self.dag.EncodeCid(path)
	c := cid.Cid{}
	err = json.Unmarshal(encodedCid.Str, &c)
	if err != nil {
		panic(fmt.Errorf("CRDTSetOpBasedDag Increment, could not UnMarshal\nerror: %s", err))
	}

	// fmt.Println("encodedCid Increment :", c.String())
	var pl CRDTDag.CRDTDagNodeInterface = &newNode

	self.dag.AddNode(encodedCid, &pl) // Adding the node created before to the Merkle-DAG

	self.SendRemoteUpdates() // Op-Based force us to send updates to other at each update

	times := TimeTuple{} // Time measurement structure, for analysis only (when self.Measurement is true)

	if self.measurement {
		//Add time
		b, err := os.ReadFile(strFile + ".timeAdd")
		if err != nil {
			panic(fmt.Errorf("Couldn't read TimeAdd file\nError: %s\n", err))
		}
		intAdd, err := strconv.Atoi(string(b))
		if err != nil {
			panic(fmt.Errorf(" TimeAdd file is malformatted, and couldn't be Atoi'ed\nError: %s\n", err))
		}
		times.Time_add = intAdd

		err = os.Remove(strFile + ".timeAdd")
		if err != nil {
			panic(fmt.Errorf("Couldn't Remove TimeAdd file\nError: %s\n", err))
		}

		// Encrypt Time
		times.Time_encrypt = 0
		if self.dag.Key != "" {
			b, err = os.ReadFile(strFile + ".timeEncrypt")
			if err != nil {
				panic(fmt.Errorf("Couldn't read timeEncrypt file\nError: %s\n", err))
			}
			intAdd, err = strconv.Atoi(string(b))
			if err != nil {
				panic(fmt.Errorf(" timeEncrypt file is malformatted, and couldn't be Atoi'ed\nError: %s\n", err))
			}
			times.Time_encrypt = intAdd

			err = os.Remove(strFile + ".timeEncrypt")
			if err != nil {
				panic(fmt.Errorf("Couldn't Remove timeEncrypt file\nError: %s\n", err))
			}
		}

	}

	return c.String(), times
}
func (self *CRDTSetOpBasedDag) Remove(x string) string {

	newNode := CreateDagNode(Operation{Elem: Element(x), Op: REMOVE}, self.GetSys().IpfsNode.Identity.ShortString())
	for dependency := range self.dag.Root_nodes {
		newNode.DagNode.DirectDependency = append(newNode.DagNode.DirectDependency, self.dag.Root_nodes[dependency])
	}

	strFile := self.dag.NextFileName()
	if _, err := os.Stat(strFile); !errors.Is(err, os.ErrNotExist) {
		os.Remove(strFile)
	}
	newNode.ToFile(strFile)
	bytes, err := os.ReadFile(strFile)
	if err != nil {
		panic(fmt.Errorf("ERROR INCREMENT CRDTSetOpBasedDag, could not read file\nerror: %s", err))
	}
	path, err := self.callAddToIPFS(bytes, strFile)
	if err != nil {
		panic(fmt.Errorf("CRDTSetOpBasedDag Decrement, could not add the file to IFPS\nerror: %s", err))
	}

	encodedCid := self.dag.EncodeCid(path)
	c := cid.Cid{}
	err = json.Unmarshal(encodedCid.Str, &c)
	if err != nil {
		panic(fmt.Errorf("CRDTSetOpBasedDag Increment, could not UnMarshal\nerror: %s", err))
	}

	// _, c, _ := cid.CidFromBytes(encodedCid.Str)
	// fmt.Println("encodedCid Decrement :", c.String())
	var pl CRDTDag.CRDTDagNodeInterface = &newNode
	self.dag.AddNode(encodedCid, &pl)
	self.SendRemoteUpdates()
	self.GetDag().UpdateRootNodeFolder()
	return c.String()
}

func Create_CRDTSetOpBasedDag(sys *IpfsLink.IpfsLink, cfg Config.CRONUSConfig) CRDTSetOpBasedDag {

	man := CRDTDag.Create_CRDTManager(sys, cfg.PeerName, cfg.BootstrapPeer, cfg.Encode, cfg.Measurement)
	crdtSet := CRDTSetOpBasedDag{dag: man, measurement: cfg.Measurement}
	if cfg.BootstrapPeer == "" {
		x, err := os.ReadFile("initial_value")
		if err != nil {
			panic(fmt.Errorf("Could not read initial_value, error : %s", err))
		}
		newNode := CreateDagNode(Operation{Elem: Element(x), Op: ADD}, crdtSet.GetSys().IpfsNode.Identity.ShortString())
		strFile := crdtSet.dag.NextFileName()

		if _, err := os.Stat(strFile); !errors.Is(err, os.ErrNotExist) {
			os.Remove(strFile)
		}
		newNode.ToFile(strFile)

		bytes, err := os.ReadFile(strFile)
		if err != nil {
			panic(fmt.Errorf("ERROR INCREMENT CRDTSetOpBasedDag, could not read file\nerror: %s", err))
		}
		path, err := man.AddToIPFS(crdtSet.dag.Sys, bytes) // Add Inital State ( so it isn't counted as messages)
		if err != nil {
			panic(fmt.Errorf("CRDTSetOpBasedDag Increment, could not add the file to IFPS\nerror: %s", err))
		}

		encodedCid := crdtSet.dag.EncodeCid(path)
		c := cid.Cid{}
		err = json.Unmarshal(encodedCid.Str, &c)
		if err != nil {
			panic(fmt.Errorf("CRDTSetOpBasedDag Increment, could not UnMarshal\nerror: %s", err))
		}
		// fmt.Println("encodedCid Increment :", c.String())
		var pl1 CRDTDag.CRDTDagNodeInterface = &newNode

		crdtSet.dag.AddNode(encodedCid, &pl1) // TODOSetCrdt Complete Node interface

	}
	var pl CRDTDag.CRDTDag = &crdtSet

	CRDTDag.CheckForRemoteUpdates(&pl, sys.Cr.Sub, man.Sys.Ctx)

	return crdtSet
}

func (self *CRDTSetOpBasedDag) GetSys() *IpfsLink.IpfsLink {

	return self.dag.Sys
}

func (self *CRDTSetOpBasedDag) Lookup_ToSpecifyType() *CRDT.CRDT {

	crdt := CRDTSetOpBased{
		sys:     self.GetSys(),
		added:   make([]string, 0),
		removed: make([]string, 0),
	}
	for x := range self.dag.GetAllNodes() {
		node := self.dag.GetAllNodesInterface()[x]
		if (*(*node).GetEvent()).(*PayloadOpBased).Op.Op == ADD {
			// fmt.Println("add")
			crdt.Add(string((*(*node).GetEvent()).(*PayloadOpBased).Op.Elem))
		} else {
			// fmt.Println("remove")
			crdt.Remove(string((*(*node).GetEvent()).(*PayloadOpBased).Op.Elem))
		}
	}
	var pl CRDT.CRDT = &crdt
	return &pl
}
func (self *CRDTSetOpBasedDag) Lookup() CRDTSetOpBased {

	// crdt := self.logokup_ToSpecifyType()
	// var pl CRDTDag.CRDTDag = &crdtSet
	return *(*self.Lookup_ToSpecifyType()).(*CRDTSetOpBased)
}

type TimeTuple struct {
	Cid            string
	RetrievalAlone int
	SeekAlone      int
	RetrievalTotal int
	SeekTotal      int
	CalculTime     int
	Time_add       int
	Time_encrypt   int
	Time_decrypt   int
	ArrivalTime    int
	FileSize       int

	TimeTotal            int
	TimeFolder           int
	TimefilesMeasurement int

	TimeselfaddCIDs               int
	TimeGetSema                   int
	Timeselfupdaterootdnodefolder int

	TimeMerge int

	TimeReadinMerge      int
	TimeGetinMerge       int
	ForloopinMerge       int
	TimeCreateDAGNODE    int
	TimeFromFile         int
	TimeremoteAddNodefor int

	CheckDependency      int
	GetNodeFromEncoded   int
	CreateNodeFromFile   int
	TimeAddNodeInCRDTDAG int
}

// semaphore usage
func getSema(sema *semaphore.Weighted, ctx context.Context) {
	t := time.Now()
	err := sema.Acquire(ctx, 1)
	for err != nil && time.Since(t) < 10*time.Second {
		time.Sleep(10 * time.Microsecond)
		err = sema.Acquire(ctx, 1)
	}
	if err != nil {
		panic(fmt.Errorf("Semaphore of READ/WRITE file locked !!!!\n Cannot acquire it\n"))
	}
}

func returnSema(sema *semaphore.Weighted) {
	sema.Release(1)
}

// Check update function retrieve files from ipfs (long)
// and then reserves the semaphore to actually modify the data (short)
func (self *CRDTSetOpBasedDag) CheckUpdate(sema *semaphore.Weighted) []TimeTuple {
	// Defineing additionnal measurement time :
	timetotal := time.Now()
	time_Files_measruement := time.Since(timetotal)
	time_folder := time.Since(timetotal)
	timeself_addCIDs := 0
	timeself_updaterootdnodefolder := 0
	time_getsema := 0

	received := make([]TimeTuple, 0)
	files, err := ioutil.ReadDir(self.GetDag().Nodes_storage_enplacement + "/remote")
	if err != nil {
		fmt.Printf("CheckUpdate - Checkupdate could not open folder\nerror: %s\n", err)
	} else {
		time_Files := time.Now()
		time_folder = time.Since(timetotal)
		ti := time.Now()
		to_add := make([]([]byte), 0)
		computetime := make([]int64, 0)
		arrivalTime := make([]int64, 0)
		for _, file := range files {
			if file.Size() > 0 && !strings.Contains(file.Name(), ".ArrivalTime") {
				fil, err := os.OpenFile(self.GetDag().Nodes_storage_enplacement+"/remote/"+file.Name(), os.O_RDONLY, os.ModeAppend)
				if err != nil {
					fmt.Printf("error in checkupdate, Could not open the sub file\nError: %s", err)
				}
				stat, err := fil.Stat()
				if err != nil {
					panic(fmt.Errorf("error in checkupdate, Could not get stat the sub file\nError: %s", err))
				}
				bytesread := make([]byte, stat.Size())
				n, err := fil.Read(bytesread)
				if err != nil {
					panic(fmt.Errorf("error in checkupdate, Could not read the sub file\nError: %s", err))
				}

				// fmt.Println("stat.size :", stat.Size(), "read :", n)
				if int64(n) != stat.Size() {
					panic(fmt.Errorf("error in checkupdate, Could not read entirely the sub file\nError: read %d byte unstead of %d", n, stat.Size()))
				}
				err = fil.Close()
				if err != nil {
					panic(fmt.Errorf("error in checkupdate, Could not close the sub file\nError: %s", err))
				}
				if !self.IsKnown(CRDTDag.EncodedStr{Str: bytesread}) {
					to_add = append(to_add, bytesread)
				}
				s := cid.Cid{}
				json.Unmarshal(bytesread, &s)

				err = os.Remove(self.GetDag().Nodes_storage_enplacement + "/remote/" + file.Name())
				if err != nil && !(errors.Is(err, os.ErrNotExist)) {
					fmt.Printf("error in checkupdate, Could not remove the sub file\nError: %s", err)
				}

				// Take the time measurement of this file
				// Get the time of arrival to compute pubsub time

				// Take the time measurement of this file
				// Get the time of arrival to compute pubsub time
				fil, err = os.OpenFile(self.GetDag().Nodes_storage_enplacement+"/remote/"+file.Name()+".ArrivalTime", os.O_RDONLY, os.ModeAppend)
				if err != nil {
					fmt.Printf("error in checkupdate, Could not open the sub file\nError: %s", err)
				}
				stat, err = fil.Stat()
				if err != nil {
					panic(fmt.Errorf("error in checkupdate, Could not get stat the sub file\nError: %s", err))
				}
				bytesread = make([]byte, stat.Size())
				n, err = fil.Read(bytesread)
				if err != nil {
					panic(fmt.Errorf("error in checkupdate, Could not read the sub file\nError: %s", err))
				}

				// filez.WriteString("5\n")
				fmt.Println("stat.size :", stat.Size(), "read :", n)
				if int64(n) != stat.Size() {
					panic(fmt.Errorf("error in checkupdate, Could not read entirely the sub file\nError: read %d byte unstead of %d", n, stat.Size()))
				}
				err = fil.Close()
				if err != nil {
					panic(fmt.Errorf("error in checkupdate, Could not close the sub file\nError: %s", err))
				}
				time_of_arrival, _ := strconv.Atoi(string(bytesread))
				arrivalTime = append(arrivalTime, int64(time_of_arrival))

				// filez.WriteString("6\n")
				//computation time, time to manage this file
				timeToCompute := time.Since(ti).Nanoseconds()
				computetime = append(computetime, timeToCompute)
				ti = time.Now()

				// 	}
				// 	if exists {
				// 		if !self.IsKnown(CRDTDag.EncodedStr{Str: bytesread}) {
				// 			to_add = append(to_add, bytesread)
				// 		}
				// 		err = os.Remove(self.GetDag().Nodes_storage_enplacement + "/remote/" + file.Name())
				// 		if err != nil || errors.Is(err, os.ErrNotExist) {
				// 			panic(fmt.Errorf("error in checkupdate, Could not remove the sub file\nError: %s", err))
				// 		}

				// 		bytesread = make([]byte, stat.Size())
				// 		n, err = fil.Read(bytesread)
				// 		if err != nil {
				// 			panic(fmt.Errorf("error in checkupdate, Could not read the sub file\nError: %s", err))
				// 		}

				// 		// fmt.Println("stat.size :", stat.Size(), "read :", n)
				// 		if int64(n) != stat.Size() {
				// 			panic(fmt.Errorf("error in checkupdate, Couldreceived not read entirely the sub file\nError: read %d byte unstead of %d", n, stat.Size()))
				// 		}
				// 		time_of_arrival, _ := strconv.Atoi(string(bytesread))
				// 		arrivalTime = append(arrivalTime, int64(time_of_arrival))

				// 		//computation time, time to manage this file
				// 		timeToCompute := time.Since(ti).Nanoseconds()
				// 		computetime = append(computetime, timeToCompute)
				// 		ti = time.Now()
				// 	} else {
				// 		fmt.Printf("File doesn't exists yet, ignoring and waiting")
				// 	}add_cids
				// }
			} else {
				fmt.Printf("Remote folder contain a FILE of a NULL SIZE\n")
			}
		}
		time_Files_measruement = time.Since(time_Files)
		// Not measured bellow (add_cids measure itself, but we dont measure outside here the semaphore)
		// apply the update on the peer's data
		if len(to_add) > 0 {
			time_get_Sema := time.Now()
			getSema(sema, self.GetSys().Ctx) //TODO : This sema might block
			time_getsema = time_getsema + int(time.Since(time_get_Sema).Nanoseconds())
			ti := time.Now()
			time_self_addCIDs := time.Now()
			received = self.add_cids(to_add, computetime, arrivalTime, ti)
			timeself_addCIDs = timeself_addCIDs + int(time.Since(time_self_addCIDs).Nanoseconds())

			time_self_updaterootnodefolder := time.Now()
			//This is the important line
			self.GetDag().UpdateRootNodeFolder()
			timeself_updaterootdnodefolder = timeself_updaterootdnodefolder + int(time.Since(time_self_updaterootnodefolder).Nanoseconds())

			returnSema(sema)

			if false { // TODO Remove and study
				additionnalCompute := time.Since(ti).Nanoseconds()
				for x := range received {
					received[x].CalculTime = received[x].CalculTime + int(additionnalCompute)
				}
			}
		}

	}

	t_total := time.Since(timetotal)
	for x := range received {
		received[x].TimeTotal = int(t_total.Nanoseconds())
		received[x].TimeFolder = int(time_folder.Nanoseconds())
		received[x].TimefilesMeasurement = int(time_Files_measruement.Nanoseconds())
		received[x].TimeselfaddCIDs = timeself_addCIDs
		received[x].Timeselfupdaterootdnodefolder = timeself_updaterootdnodefolder
		received[x].TimeGetSema = time_getsema
	}

	return received
}

func checkFileExists(filePath string) bool {
	_, error := os.Stat(filePath)
	//return !os.IsNotExist(err)
	return !errors.Is(error, os.ErrNotExist)
}

func Index(stringmap []([]byte), elem []byte) int {
	for i, x := range stringmap {
		if string(x) == string(elem) {
			return i
		}
	}
	return -1

}

// Check update function retrieve files from ipfs (long)
// and then reserves the semaphore to actually modify the data (short)
func (self *CRDTSetOpBasedDag) CheckUpdate_20Version(sema *semaphore.Weighted) []TimeTuple {
	fileRead, err := os.OpenFile(self.GetDag().Nodes_storage_enplacement+"/time/FileBIS.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		fmt.Printf("CheckUpdate - Cannot create FileBIS: %s\n", err)
	}
	received := make([]TimeTuple, 0)
	files, err := ioutil.ReadDir(self.GetDag().Nodes_storage_enplacement + "/remote")
	if err != nil {
		fmt.Printf("CheckUpdate - Checkupdate could not open folder\nerror: %s\n", err)
	} else {
		timelookup := time.Now()
		ti := time.Now()
		to_add := make([]([]byte), 0)
		computetime := make([]int64, 0)
		arrivalTime := make([]int64, 0)
		cpt := 0
		fileRead.WriteString("===================================================================================================\n")
		fileRead.WriteString("Starting to read the folder\n")
		willBeAdded := make([]([]byte), 0)
		for _, file := range files {
			if file.Size() > 0 && !strings.Contains(file.Name(), ".ArrivalTime") && checkFileExists(self.GetDag().Nodes_storage_enplacement+"/remote/"+file.Name()+".ArrivalTime") {
				fil, err := os.OpenFile(self.GetDag().Nodes_storage_enplacement+"/remote/"+file.Name(), os.O_RDONLY, os.ModeAppend)
				if err != nil {
					panic(fmt.Errorf("error in checkupdate, Could not open the sub file\nError: %s", err))
				}
				stat, err := fil.Stat()
				if err != nil {
					panic(fmt.Errorf("error in checkupdate, Could not get stat the sub file\nError: %s", err))
				}
				bytesread := make([]byte, stat.Size())
				n, err := fil.Read(bytesread)
				if err != nil {
					panic(fmt.Errorf("error in checkupdate, Could not read the sub file\nError: %s", err))
				}

				// fmt.Println("stat.size :", stat.Size(), "read :", n)
				if int64(n) != stat.Size() {
					panic(fmt.Errorf("error in checkupdate, Could not read entirely the sub file\nError: read %d byte unstead of %d", n, stat.Size()))
				}
				err = fil.Close()
				if err != nil {
					panic(fmt.Errorf("error in checkupdate, Could not close the sub file\nError: %s", err))
				}
				if !self.IsKnown(CRDTDag.EncodedStr{Str: bytesread}) && Index(willBeAdded, bytesread) == -1 {
					cpt = cpt + 1
					willBeAdded = append(willBeAdded, bytesread)

				}
				if cpt >= 40 {
					break
				}
			}
		}

		fileRead.WriteString(fmt.Sprintf("after reading files, cpt value is %d\n", cpt))
		if cpt >= 40 {
			time_lookup_folder := time.Since(timelookup)
			time_remove_files := time.Since(time.Now()) // init at 0
			fileRead.WriteString(fmt.Sprintf("Time Lookup Folder : %d, time Since %d \n", time_lookup_folder, time.Since(timelookup)))
			cpt = 0
			fileRead.WriteString(fmt.Sprintf("we're in \n"))
			for _, file := range files {
				if file.Size() > 0 && !strings.Contains(file.Name(), ".ArrivalTime") && checkFileExists(self.GetDag().Nodes_storage_enplacement+"/remote/"+file.Name()+".ArrivalTime") {
					fil, err := os.OpenFile(self.GetDag().Nodes_storage_enplacement+"/remote/"+file.Name(), os.O_RDONLY, os.ModeAppend)
					if err != nil {
						panic(fmt.Errorf("error in checkupdate, Could not open the sub file\nError: %s", err))
					}
					stat, err := fil.Stat()
					if err != nil {
						panic(fmt.Errorf("error in checkupdate, Could not get stat the sub file\nError: %s", err))
					}
					bytesread := make([]byte, stat.Size())
					n, err := fil.Read(bytesread)
					if err != nil {
						panic(fmt.Errorf("error in checkupdate, Could not read the sub file\nError: %s", err))
					}

					// fmt.Println("stat.size :", stat.Size(), "read :", n)
					if int64(n) != stat.Size() {
						panic(fmt.Errorf("error in checkupdate, Could not read entirely the sub file\nError: read %d byte unstead of %d", n, stat.Size()))
					}
					err = fil.Close()
					if err != nil {
						panic(fmt.Errorf("error in checkupdate, Could not close the sub file\nError: %s", err))
					}
					if !self.IsKnown(CRDTDag.EncodedStr{Str: bytesread}) {
						cpt = cpt + 1
						to_add = append(to_add, bytesread)
					}
					s := cid.Cid{}
					json.Unmarshal(bytesread, &s)

					timelookup = time.Now()
					err = os.Remove(self.GetDag().Nodes_storage_enplacement + "/remote/" + file.Name())
					if err != nil || errors.Is(err, os.ErrNotExist) {
						panic(fmt.Errorf("error in checkupdate, Could not remove the sub file\nError: %s", err))
					}
					time_remove_files = time_remove_files + time.Since(timelookup)
					// Take the time measurement of this file
					// Get the time of arrival to compute pubsub time
					fil, err = os.OpenFile(self.GetDag().Nodes_storage_enplacement+"/remote/"+file.Name()+".ArrivalTime", os.O_RDONLY, os.ModeAppend)
					if err != nil {
						panic(fmt.Errorf("error in checkupdate, Could not open the sub file\nError: %s", err))
					}
					stat, err = fil.Stat()
					if err != nil {
						panic(fmt.Errorf("error in checkupdate, Could not get stat the sub file\nError: %s", err))
					}
					bytesread = make([]byte, stat.Size())
					n, err = fil.Read(bytesread)
					if err != nil {
						panic(fmt.Errorf("error in checkupdate, Could not read the sub file\nError: %s", err))
					}

					// fmt.Println("stat.size :", stat.Size(), "read :", n)
					if int64(n) != stat.Size() {
						panic(fmt.Errorf("error in checkupdate, Could not read entirely the sub file\nError: read %d byte unstead of %d", n, stat.Size()))
					}
					err = fil.Close()
					if err != nil {
						panic(fmt.Errorf("error in checkupdate, Could not close the sub file\nError: %s", err))
					}
					time_of_arrival, _ := strconv.Atoi(string(bytesread))
					arrivalTime = append(arrivalTime, int64(time_of_arrival))

					//computation time, time to manage this file
					timeToCompute := time.Since(ti).Nanoseconds()
					computetime = append(computetime, timeToCompute)
					ti = time.Now()

					if cpt >= 40 {
						break
					}

				} else {
					fmt.Printf("Remote folder contain a FILE of a NULL SIZE\n")
				}
			}
			// apply the update on the peer's data
			getSema(sema, self.GetSys().Ctx)
			received = self.add_cids(to_add, computetime, arrivalTime, ti)

			if len(to_add) > 0 {
				self.GetDag().UpdateRootNodeFolder()
			}
			fileRead.WriteString(fmt.Sprintf("Returning %d values\n", len(to_add)))

			returnSema(sema)

			// for i := range received {
			// 	received[i].TimeLookupFolder = int(time_lookup_folder.Nanoseconds())
			// 	received[i].TimeRemoveFiles = int(time_remove_files.Nanoseconds())
			// }

			fileRead.WriteString(fmt.Sprintf("Time Remove files from Folder : %d _ value writen : %d \n", time_remove_files, int(time_remove_files.Nanoseconds())))
			fileRead.WriteString(fmt.Sprintf("Repeat time Lookup Folder : %d, time written : %d \n", time_lookup_folder, int(time_remove_files.Nanoseconds())))
		}
	}
	fileRead.WriteString(fmt.Sprintf("Returning %d files\n", len(received)))
	fileRead.Close()
	return received
}

func (self *CRDTSetOpBasedDag) add_cids(to_add []([]byte), computetime []int64, arrivalTime []int64, ti time.Time) []TimeTuple {
	received := make([]TimeTuple, 0)

	bytes_encoded := make([]CRDTDag.EncodedStr, 0)

	for _, bytesread := range to_add {
		bytes_encoded = append(bytes_encoded, CRDTDag.EncodedStr{Str: bytesread})
	}
	timeMerge := time.Now()
	// , times
	// filesWritten, cidReceived, times := self.Merge(bytes_encoded)
	filesWritten, cidReceived := self.Merge(bytes_encoded)
	timeSinceMerge := time.Since(timeMerge)

	bytes_cids := make([]CRDTDag.EncodedStr, 0)

	for _, bytesread := range cidReceived {
		bytes_cids = append(bytes_cids, CRDTDag.EncodedStr{Str: bytesread})
	}

	for index, bytesread := range cidReceived {
		correspondingFile := filesWritten[index]
		s := cid.Cid{}
		json.Unmarshal(bytesread, &s)
		timeRetrieve := 0
		timeSeek := 0
		timeDecrypt := 0
		fileSize := 0
		if self.measurement && correspondingFile != "node1/node1" {
			// Get Time of Retrieval
			str, err := os.ReadFile(correspondingFile + ".timeRetrieve")
			fileInfo, _ := os.Stat(correspondingFile)
			fileSize = int(fileInfo.Size())
			if err != nil {
				panic(fmt.Errorf("set.go - could not read time to retrieve measurement\nerror: %s", err))
			}
			timeRetrieve, err = strconv.Atoi(string(str))
			if err != nil {
				panic(fmt.Errorf("set.go - could not translate time to retrieve to string, maybe malformerd ?\nerror: %s", err))
			}

			err = os.Remove(correspondingFile + ".timeRetrieve")
			if err != nil {
				panic(fmt.Errorf("set.go - could not remove time to retrieve file\nerror: %s", err))
			}

			// Get Time of seektime
			str, err = os.ReadFile(correspondingFile + ".timeSeek")
			fileInfo, _ = os.Stat(correspondingFile)
			fileSize = int(fileInfo.Size())
			if err != nil {
				panic(fmt.Errorf("set.go - could not read time to retrieve measurement\nerror: %s", err))
			}
			timeSeek, err = strconv.Atoi(string(str))
			if err != nil {
				panic(fmt.Errorf("set.go - could not translate time to retrieve to string, maybe malformerd ?\nerror: %s", err))
			}

			err = os.Remove(correspondingFile + ".timeSeek")
			if err != nil {
				panic(fmt.Errorf("set.go - could not remove time to retrieve file\nerror: %s", err))
			}

			//If we use it, get time of decryption of the file
			if self.dag.Key != "" {
				str, err = os.ReadFile(correspondingFile + ".timeDecrypt")
				if err != nil {
					panic(fmt.Errorf("set.go - could not read time decrypt measurement\nerror: %s", err))
				}
				timeDecrypt, err = strconv.Atoi(string(str))
				if err != nil {
					panic(fmt.Errorf("set.go - could not translate time to retrieve to string, maybe malformerd ?\nerror: %s", err))
				}
				err = os.Remove(correspondingFile + ".timeDecrypt")
				if err != nil {
					panic(fmt.Errorf("set.go - could not remove time to decrypt file\nerror: %s", err))
				}

			}

		}
		// fmt.Println("calling UpdateRootNodeFolder")
		comptime := 0
		arrtime := 0
		if index < len(computetime) {
			comptime = int(computetime[index])
		} else {
			comptime = int(computetime[len(computetime)-1])

		}

		if index < len(arrivalTime) {
			arrtime = int(arrivalTime[index])
		} else {
			arrtime = int(arrivalTime[len(arrivalTime)-1])

		}

		received = append(received, TimeTuple{
			Cid:            s.String(),
			RetrievalAlone: timeRetrieve,
			RetrievalTotal: timeRetrieve * len(to_add),
			SeekAlone:      timeSeek,
			SeekTotal:      timeSeek * len(to_add),
			CalculTime:     comptime,
			ArrivalTime:    arrtime,
			Time_decrypt:   timeDecrypt,
			Time_encrypt:   0,
			FileSize:       fileSize,
			TimeMerge:      int(timeSinceMerge.Nanoseconds()),

			// TimeReadinMerge:      times.TimeReadinMerge,
			// TimeGetinMerge:       times.TimeGetinMerge,
			// ForloopinMerge:       times.ForloopinMerge,
			// TimeCreateDAGNODE:    times.TimeCreateDAGNODE,
			// TimeFromFile:         times.TimeFromFile,
			// TimeremoteAddNodefor: times.TimeremoteAddNodefor,

			// CheckDependency:      times.Timings.CheckDependency,
			// GetNodeFromEncoded:   times.Timings.GetNodeFromEncoded,
			// CreateNodeFromFile:   times.Timings.CreateNodeFromFile,
			// TimeAddNodeInCRDTDAG: times.Timings.TimeAddNodeInCRDTDAG,
		})

	}
	self.GetDag().UpdateRootNodeFolder()
	return received
}
