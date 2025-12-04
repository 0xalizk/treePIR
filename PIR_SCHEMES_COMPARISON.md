# PIR Schemes Comparison: TreePIR vs Punc vs DPF

This document explains the key differences between the three main PIR (Private Information Retrieval) schemes implemented in this codebase.

## Overview

All three schemes are **two-server PIR protocols** that allow a client to privately retrieve a database record without revealing which record they're interested in. They differ in their underlying cryptographic primitives, performance characteristics, and trade-offs.

---

## 1. **DPF (Distributed Point Function)**

### What it is
DPF is based on **Function Secret Sharing** (Boyle, Gilboa, and Ishai, 2018). It uses Distributed Point Functions to create compact queries.

### Key Characteristics

**Offline Phase:**
- **Server time**: 0 (no preprocessing)
- **Communication**: 0 (no hints sent)
- **Client storage**: 0

**Online Phase:**
- **Server time**: O(n) - servers must process all n database rows
- **Communication**: O(λ log n) - very small query size (just a DPF key)
- **Query size**: ~λ log(n) bits (typically a few KB)

### How it Works
1. Client generates two DPF keys that encode a point function (1 at the desired index, 0 elsewhere)
2. Each server evaluates its DPF key to get a bit vector
3. Servers compute matrix-vector product: XOR all database rows where bit vector is 1
4. Client XORs the two responses to recover the desired row

### Implementation
- Uses optimized C++ code for DPF evaluation
- No offline preprocessing required
- Very efficient for small queries

### Best For
- **One-time queries** (no offline phase overhead)
- **Low communication** requirements
- **Simple deployment** (no hint management)

### Trade-offs
- **High server computation**: Must process entire database for each query
- **No offline optimization**: Can't amortize work across queries

---

## 2. **Punc (Puncturable Sets)**

### What it is
Punc is Checklist's **offline/online PIR scheme** from the original Checklist paper (Kogan & Corrigan-Gibbs, 2021). It uses **puncturable sets** to enable efficient online queries after offline preprocessing.

### Key Characteristics

**Offline Phase:**
- **Server time**: O(λn) - generates hints by XORing sets of rows
- **Communication**: O(λ√n) - sends hints to client
- **Client storage**: O(λ√n) - stores hints locally
- **Set size**: √n (square root of database size)

**Online Phase:**
- **Server time**: O(√n) - only processes a subset of rows
- **Communication**: O(λ log n) - small query size
- **Query size**: Punctured set + extra element

### How it Works
1. **Offline**: Server generates many random sets of size √n, XORs their rows to create hints
2. **Offline**: Client stores these hints
3. **Online**: Client finds a hint containing desired index, punctures it (removes the index)
4. **Online**: Client sends two punctured sets to servers (with different extra elements)
5. **Online**: Servers XOR rows in the punctured sets and return results
6. **Online**: Client reconstructs desired row from responses

### Implementation
- Uses puncturable set generator (PSet)
- Requires offline preprocessing
- Optimized C++ code for set operations

### Best For
- **Multiple queries** (amortizes offline cost)
- **Balanced performance** (good online time + communication)
- **Updatable databases** (supports efficient updates)

### Trade-offs
- **Offline overhead**: Requires preprocessing
- **Client storage**: Must store hints
- **Moderate complexity**: More complex than DPF

---

## 3. **TreePIR**

### What it is
TreePIR is a **newer scheme** based on the TreePIR paper (2023). It's an evolution of Punc that uses **tree-based puncturable sets** for potentially better performance.

### Key Characteristics

**Offline Phase:**
- **Server time**: O(λn) - similar to Punc
- **Communication**: O(λ√n) - similar to Punc  
- **Client storage**: O(λ√n) - similar to Punc
- **Set size**: √n (same as Punc)

**Online Phase:**
- **Server time**: O(√n) - similar to Punc
- **Communication**: O(λ log n) - similar to Punc
- **Query size**: Tree-based punctured set + extra element

### How it Works
1. Uses **tree-based puncturable sets** (weak privately puncturable PRF-based sets)
2. Similar structure to Punc but with different set generation/evaluation
3. Uses `GenWithTwo()` and `PuncTwo()` instead of `GenWith()` and `Punc()`
4. Tree structure may provide better locality and performance

### Implementation
- Uses optimized C++ implementation in `psetggm/`
- Tree-based set generation (`NewSetGeneratorTwo`)
- Height array precomputation for efficiency
- Similar to Punc but with tree optimizations

### Best For
- **Modern implementations** (latest research)
- **Potentially better performance** than Punc (tree optimizations)
- **Same use cases as Punc** (multiple queries, updatable)

### Trade-offs
- **Similar to Punc** in complexity and overhead
- **Newer scheme** (may have less real-world testing)
- **Tree optimizations** may provide better cache locality

---

## Performance Comparison Summary

For a database of size **n** rows, with security parameter **λ ≈ 128**:

| Scheme | Offline Server Time | Offline Comm | Online Server Time | Online Comm | Client Storage |
|--------|-------------------|--------------|-------------------|-------------|----------------|
| **DPF** | 0 | 0 | **O(n)** | O(λ log n) | 0 |
| **Punc** | O(λn) | O(λ√n) | **O(√n)** | O(λ log n) | O(λ√n) |
| **TreePIR** | O(λn) | O(λ√n) | **O(√n)** | O(λ log n) | O(λ√n) |

### Key Differences

1. **DPF**: No offline phase, but must process entire database online
2. **Punc/TreePIR**: Offline preprocessing, but much faster online queries (only √n work)

### When to Use Which?

- **Use DPF** if:
  - You have very few queries (can't amortize offline cost)
  - You want zero client storage
  - Server computation is cheap
  - Communication must be minimal

- **Use Punc** if:
  - You have many queries (amortize offline cost)
  - You want fast online queries
  - You can store hints on client
  - You need updatable databases

- **Use TreePIR** if:
  - Same as Punc, but want latest optimizations
  - You want tree-based improvements
  - You're doing research/comparison

---

## Code Locations

- **DPF**: `pir/pir_dpf.go` + `modules/dpf-go/`
- **Punc**: `pir/pir_punc.go` + `pir/pset.go`
- **TreePIR**: `pir/pir_punc_tree.go` + `psetggm/pset_ggm.cpp`

## References

- **DPF**: [Function Secret Sharing](https://eprint.iacr.org/2018/707) (Boyle, Gilboa, Ishai)
- **Punc**: [Checklist Paper](https://eprint.iacr.org/2021/345.pdf) (Kogan & Corrigan-Gibbs)
- **TreePIR**: [TreePIR Paper](https://eprint.iacr.org/2023/204) (2023)


