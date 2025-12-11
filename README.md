# **CRONUS - Merkle-CRDTs over IPFS Implementation & Experimental Framework**

This repository contains a Go implementation of **Merkle-CRDTs** built on top of **IPFS Kubo v0.37.0**, including several CRDT definitions, an IPFS-based exchange layer, and the complete experimental setup used in the evaluation section of our scientific paper.

The project aims to demonstrate how Merkle-CRDTs can serve as a foundation for **fully peer-to-peer collaborative applications**, using IPFS for content addressing, and strong availability of file, and using pubsub to announce the files.

The code and scripts published here enable reviewers and researchers to:
- Inspect the full implementation used in the paper  
- Reproduce experiments (locally or on a distributed testbed)  
- Extend or adapt the CRDT definitions  
- Evaluate alternative network conditions or update patterns  

---

## **Repository Overview**
```
CRDT_IPFS/
  ├── Crdt/                 # Generic Merkle-CRDT interface and core logic
  ├── CRDTDag/              # Merkle DAG implementation for CRDT states
  ├── example/              # CRDT examples & test harnesses
  │   ├── 2PSet/
  │   ├── CLSet/
  │   ├── CLSetDelta/
  │   ├── NoConcurrency/    # IPFS-only comparison baseline
  │   └── tests/            # Experiment entry points (invoked via main.go)
  │   ├── Counter/          # Work-in-progress CRDT
  │   ├── LogootOpBased/    # Work-in-progress CRDT
  ├── ipfsLink/             # IPFS wrapper (DAG ops, publish/subscribe, connections)
  ├── Payload/              # Data structures stored as CID-addressed payloads
  ├── Config/               # Experiment configuration options
  ├── g5kRunners/           # Helpers for running on Grid5000
  ├── main.go               # Experiment launcher selecting the scenario
  ├── TexFiles/             # Paper fragments and generated PDFs
  └── tuto_IPFS_Deployment.md

```

Other top-level folders:

- **General Workflow/** — Draw.io diagrams and visuals used in the work  
- **ScriptExperiment/** — Shell scripts and R tools for running/collecting results on Grid5000  
- **toremove/kad/** — Vendored copy of go-libp2p-kad-dht (kept for reproducibility)  
- **output.csv**, **latencyTime_new.ods** — Example outputs  

---

## **Implemented CRDTs**

The `example` folder contains several CRDT definitions integrated with the Merkle-CRDT framework:

### **Stable CRDT Implementations**
- **2P-Set**  
- **CL-Set (state-based)**  
- **Delta-CL-Set (delta-based version)**  
- **Empty / template example** -- As a simple example on how to define any CRDT

### **Work-in-Progress (experimental)**
- **Counter** (simple grow-only counter)
- **Logoot-Op-Based** structure  

### **Comparison baseline**
- **ipfs_only.go** in `example/NoConcurrency/`  
  A mode that bypasses CRDT logic and simply shares full files through IPFS.  

  Used as a baseline in experiment comparisons, this cannot handle any concurrent update and has no system to prevent concurrent update from happening, only use with 1 updater.

---

## **Experimental Entry Points**

Although individual CRDTs live in separate folders, **all experiments are launched via `example/tests`**.  
These files behave as main entry points depending on configuration.

Examples include:
- `Comparison_CRDT.go`  (For 2P set)
- `Comparison_CRDT_StateBased.go`  
- `Comparison_CRDT_DeltaBased.go`  
- `Comparison_CRDT_LogootOpBased.go`  
- `Comparison_CRDT_NoCRDT.go`  
- `Sharing_Files_test.go`  

The root `main.go` file loads parameters and dispatches execution to these test modules.

---

## **IPFS Integration**

Communication layer is separated within two complementary tools:

### **1. IPFS  -  Content addressing & file share**
Located in:  

```
CRDT_IPFS/ipfsLink/
   ├── Client.go
   └── IpfsLink.go

```
This includes:
- Publication of CRDT Payloads (i.e. State or update)
- Retrieval of Files while knowing only its Identifier (CID)

### **2. libp2p PubSub**
Used to broadcast updates and coordinate synchronization events across peers, also located in :

```
CRDT_IPFS/ipfsLink/
   └── IpfsLink.go

```
### **Supported IPFS Version**
The framework is built on **IPFS Kubo v0.37.0**.  
Using other versions (after v0.30.0) may work but is not experimented yet.
       
---

## **Running Experiments Locally**

### **Prerequisites**
- Go ≥ 1.21
- IPFS Kubo v0.37.0 go implementation
- IPFS configured for pubsub and local discovery (see `tuto_IPFS_Deployment.md`)

### **Build**

```
--sh
cd CRDT_IPFS
go mod tidy
go build 
```

### Run 

Most experiments are controlled through the experiment's configuration data "config.cfg" that is set-up thanks to flags defin :
- TODO Describe the flags


Usage :
```
 TODO
```

Configuration options determine:
- The CRDT to use
- Number of peers
- Update patterns
- Delays / timers
- Storage locations
- Debug vs. experiment mode

## **Running Experiments on Grid5000 (G5K)**

The `ScriptExperiment` folder includes all scripts used to deploy nodes, control execution, and collect results on the Grid5000 testbed.

Key components: (TODO)  


These scripts are provided for transparency and reproducibility.

---

## **Extending the Framework**

This framework is designed for people wishing to develop or benchmark new CRDTs on IPFS.

### To implement a new CRDT:
1. Create a new folder inside `example/`
2. Implement: (TODO, mention the explaination)
    
3. Add an experiment entry point in `example/tests/`
4. Configure the corresponding experiment parameters

The architecture is modular and intended to facilitate experimentation and reproducibility.

## **Contact**

This repository is anonymized for the review process.  
Contact information will be added after the review period.
