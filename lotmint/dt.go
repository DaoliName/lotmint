/* 3.2 Global Clock Cycle Mapping to/from Private Clock Cycles中提到的时间映射算法 */

/*
package lotmint


const (
    W = 11
    I = 2
)

func GetCurrentTime() uint64_t {
    return GetTimeMillis();
}

func CalcGlobalClockCycle(const CBlockIndex *pindexNew) {
    // GC_Cycle(TB) = (timestamp(TB) − timestamp(TB − w + 1 − i)) / (w + i)
    // CBlockIndex以Bitcoin数据结构为例，后续使用LotMint替换
    const CBlockIndex* pindex = pindexNew;
    uint64_t totalTime = 0;
    for (int i = 0; i < W - 1 + I && pindex != nullptr; i++)
    {
        totalTime += pindex->GetBlockTime() - pindex->pprev->GetBlockTime();
        pindex = pindex->pprev;
    }
    uint64_t medianTime = ((double) totalTime) / (W + I);
    return medianTime;
}

func CalcPrivateClockCycle(const PrivateClock *privateClock) uint64_t {
    // PrivateCycles_N(TB) = (PrivateClock_N(TB) − PrivateClock_N(TB − w + 1 − i)) / (w + i)
    const PrivateClock* pClock = privateClock;
    uint64_t totalTime = 0;
    for (int i = 0; i < W - 1 + I && pClock != nullptr; i++)
    {
        totalTime += pClock->GetBlockTime() - pClock->pprev->GetBlockTime();
        pClock = pClock->pprev;
    }
    uint64_t medianTime = ((double) totalTime) / (W + I);
    return medianTime;
}

func CalcDelta(CBlockIndex *pindexNew, PrivateClock *privateClock, uint64_t evtPrivateClock) uint64_t {
    // δ_N(TB,Evt) = (GC_Cycle(TB) / PrivateCycles_N(TB)) * (PrivateClock_N(Evt) − PrivateClock_N(TB))
    uint64_t delta = (CalcGlobalClockCycle(pindexNew) / CalcPrivateClockCycle(pClock)) * (evtPrivateClock - pindexNew->GetBlockTime())
    return delta;
}
*/
