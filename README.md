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
  ├── Config/               # Experiment configuration options definition
  ├── main.go               # Experiment launcher selecting the scenario
  └── tuto_IPFS_Deployment.md

```

Other top-level folders:

- **ScriptExperiment/** — Shell scripts and R tools for running/collecting results on Grid5000
- send_files.sh input —  compress CRDT_IPFS and send it to input

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


### **Detailled used algorithm implementation**


The CRDT definitions that are being used in \sys are 2P-Set for Operation-based, CLSet for State-based and $\delta$-CLSet for Delta-based. The implementations used are being depicted in Algorithm 1, 2 and 3. For clarity, the set elements are treated as strings, and uniqueness is ensured by attaching a timestamp and the peer identifier to each element.

__<u>Algorithm 1 : Operation-based 2P-Set</u>__

    Requirements:
      (S_A, S_R) ∈ String Set × String Set
      u ∈ {(t, s) | t ∈ {add, remove}, s ∈ String}

    Procedure ManageUpdate(S, u):
      (t, s) ← u

      if (t == add) then
        S_A ← S_A ∪ {s}
      else if (t == remove) then
        S_R ← S_R ∪ {s}
      end if

      SendPayload(json(u))

    Procedure SendPayload(m):
      Cid ← IPFS.Send(Merkle-CRDTNode(m))
      Pubsub.emit(Cid)



Operation-based 2P-Set uses two subsets to represent the sets $S_A$ and $S_R$. All added elements are stored in $S_A$, and all removed element are stored in $S_R$. The final set is computed as $S = S_A \setminus S_R$. The payload created while adding or removing elements is \texttt{Add x} if $x$ is being added, or  \texttt{Remove x} if $x$ is being removed. The peers will automatically compute it as $S_A = S_A \cup \{x\}$ or $S_R = S_R \cup \{x\}$ .

__<u>Algorithm 2 : CLSet update in state-based CRDT</u>__

    Requirements:
      S : Map<String, Int>

    Procedure ManageUpdate(S, u):
      if (u.type == add) then
        if (S[u.string] is even) then
          S[u.string] ← S[u.string] + 1
        Pubsub.emit(Cid)
        S_delta ← {}
      end if
      

In the delta-based $\delta$-CLSet, the data is represented in the same way as in the state-based version, but with an additional map $\mathcal{M}_\delta$. The add and remove operations behave identically; however, whenever the main map $\mathcal{M}$ is updated for a string $x$, the delta map is also updated so that $\mathcal{M}_\delta[x] = \mathcal{M}[x]$ in the sets $S_
        end if
      else if (u.type == remove) then
        if (S[u.string] is odd) then
          S[u.string] ← S[u.string] + 1
        end if
      end if

    Procedure SendPayload(S):
      if (S has been modified since last time) then
        Cid ← IPFS.Send(Merkle-CRDTNode(S.State))
        Pubsub.emit(Cid)
      end if



In the state-based approach, CLSet represents its data using a map $\mathcal M : \texttt{String} \mapsto \mathcal{N}$. An element is considered absent from the node if the number associated with its string is ev
        Pubsub.emit(Cid)
        S_delta ← {}
      end if
      

In the delta-based $\delta$-CLSet, the data is represented in the same way as in the state-based version, but with an additional map $\mathcal{M}_\delta$. The add and remove operations behave identically; however, whenever the main map $\mathcal{M}$ is updated for a string $x$, the delta map is also updated so that $\mathcal{M}_\delta[x] = \mathcal{M}[x]$ in the sets $S_en, and present if the number is odd. When a peer adds or removes an element, it simply increments the corresponding value in the map by one. The state-based payload transmits the entire map, and the merge operation takes the maximum value for each entry.

__<u>Algorithm 3 : CLSet in Delta-based CRDT</u>__

    Requirements:
      S       : Map<String, Int>
      S_delta : Map<String, Int>

    Procedure ManageUpdate(S, u):
      (t, s) ← u

      if (u.type == add) then
        if (S[s] is even) then
          S[s] ← S[s] + 1
          S_delta[s] ← S[s]
        end if
      else if (u.type == remove) then
        if (S[s] is odd) then
          S[s] ← S[s] + 1
          S_delta[s] ← S[s]
        end if
      end if

    Procedure SendPayload(S):
      if (S_delta != {}) then
        Cid ← IPFS.Send(Merkle-CRDTNode(S_delta))
        Pubsub.emit(Cid)
        S_delta ← {}
      end if
      

In the delta-based $\delta$-CLSet, the data is represented in the same way as in the state-based version, but with an additional map $\mathcal{M}_\delta$. The add and remove operations behave identically; however, whenever the main map $\mathcal{M}$ is updated for a string $x$, the delta map is also updated so that $\mathcal{M}_\delta[x] = \mathcal{M}[x]$ in the sets $S_A$ and $S_R$. The payload then transmits only $\mathcal{M}_\delta$, and after each transmission, $\mathcal{M}_\delta$ is reset to $\emptyset$.


## **Experimental Entry Points**

Although individual CRDTs live in separate folders, **all experiments are launched via `example/tests`**.  
These files behave as main entry points depending on configuration.

Examples include:
- `Comparison_CRDT.go`  (For 2P set - Operation-based)
- `Comparison_CRDT_StateBased.go`  
- `Comparison_CRDT_DeltaBased.go`  
- `Comparison_CRDT_LogootOpBased.go`  
- `Comparison_CRDT_NoCRDT.go`  (For IPFS only experiment)

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

Most experiments are controlled through the experiment's configuration data "config.cfg" that is set-up thanks to its flags 


Configuration options determine:
- Number of peers
- Update patterns
- Delays / timers
- Storage locations
- Debug vs. experiment mode

The CRDT to use is defined with the function being called in `main.go` 

## **Running Experiments on Grid5000 (G5K)**

The `ScriptExperiment` folder includes all scripts used to deploy nodes, control execution, and collect results on the Grid5000 testbed.

File I use to start experiment, calling the other files: `ScriptExperiment/run_multiple.sh`
This file requires a 


These scripts are provided for transparency and reproducibility.

---

## **Extending the Framework**

This framework is designed for people wishing to develop or benchmark new CRDTs on IPFS.

### To implement a new CRDT:
1. Create a new folder inside `example/`
2. Implement: follow the template provided in example "empty"
    
3. Add an experiment entry point in `example/tests/`
4. Configure the corresponding experiment parameters

The architecture is modular and intended to facilitate experimentation and reproducibility.

## **Contact**

This repository is anonymized for the review process.
