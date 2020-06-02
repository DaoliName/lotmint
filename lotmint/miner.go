/*
package miming

import (
    "runtime"
    "sync"
)

const (
    hashUpdateSecs = 15
)

var (
    // defaultNumWorkers is the default number of workers to use for mining
    // and is based on the number of processor cores. This helps ensure the
    // system stays reasonably responsive under heavy load.
    defaultNumWorkers = uint32(runtime.NumCPU())
)

// Miner provides facilities for solving blocks (mining) using the CPU in
// a concurrency-safe manner. It consists of a worker goroutines which
// generate and solve blocks. The number of goroutines can be set via the
// SetMaxGoRoutines function, but the default is based on the number of
// processor cores in the system which is typically sufficient.
type Miner struct {
    sync.Mutex
    numWorkers          uint32
    started             bool
    wg                  sync.WaitGroup
    workerWg            sync.WaitGroup
    updateNumWorkers    chan struct{}
    quit                chan struct{}
}

// Start begins the CPU mining process. Calling this function when the
//CPU miner has already been started will have no effect.
func (m *Miner) Start() {
    m.Lock()
    defer m.Unlock()

    if m.started {
        return
    }

    m.quit = make(chan struct{})
    m.wg.Add(1)
    go m.miningWorker()

    m.started = true
    log.Infof("CPU miner started")
}

// Stop gracefully stops the mining process by signalling all workers to quit.
// Calling this function when the CPU miner has not already been started will
// have no effect.
// This function is safe for concurrent access.
func (m *Miner) Stop() {
    m.Lock
    defer m.Unlock()

    if !m.started {
        return
    }

    close(m.quit)
    m.wg.Wait()
    m.started = false
    log.Infof("CPU miner stopped")
}

// SetNumWorkers sets the number of workers to create which solve blocks. Any
// negative values will cause a default number of workers to be used which is
// based on the number of processor cores in the system. A value of 0 will
// cause all CPU mining to be stopped.
//
// This function is safe for concurrent access.
func (m *Miner) SetNumWorkers(numWorkers int32) {
    if numWorkers == 0 {
        m.Stop()
    }

    // Don't lock until after the first check since Stop does its won locking.
    m.Lock()
    defer m.Unlock()

    // Use default if provided value is negative
    if numWorkers < 0 {
        m.numWorkers = defaultNumWorkers
    } else {
        m.numWorkers = uint32(numWorkers)
    }

    // When the miner is already running, notify the controller about the change.
    if m.started {
        m.updateNumWorkers <- struct{}{}
    }
}

// miningWorker launches the worker goroutines that are used to generate block
// templates and solve them. It also provides the ability to dynamically adjust
// the number of running worker goroutines.
//
// It must be run as a goroutine.
func (m *Miner) miningWorker() {
    // launchWorkers groups common code to launch a specified number of
    // workers for generating blocks.
    var runningWorkers []chan struct{}
    launchWorkers := func(numWorkers uint32) {
        for i := uint32(0); i < numWorkers; i++ {
            quit := make(chan struct{})
            runningWorkers = append(runningWorkers, quit)

            m.workerWg.Add(1)
            go m.generateBlocks(quit)
        }
    }

    // Launch the current number of workers by default.
    runningWorkers = make([]chan struct{}, 0, m.numWorkers)
    launchWorkers(m.numWorkers)

out:
    for {
        select {
        // Update the number of running workers.
        case <-m.updateNumWorkers:
            // No change.
            numRunning := uint32(len(runningWorkers))
            if m.numWorkers == numRunning {
                continue
            }

            // Add new workers.
            if m.numWorkers > numRunning {
                launchWorkers(m.numWorkers - numRunning)
                continue
            }

            // Signal the most recently created goroutines to exit.
            for i := numRunning - 1; i >= m.numWorkers; i-- {
                close(runningWorkers[i])
                runningWorkers[i] = nil
                runningWorkers = runningWorkers[:i]
            }
        case <-m.quit:
            for _, quit := range runningWorkers {
                close(quit)
            }
            break out
        }
    }

    // Wait until all workers shut down.
    m.workerWg.Wait()
    m.wg.Done()
}

// generateBlocks is a worker that is controlled by the miningWorker. It is
// self contained in that it creates block templates and attempts to solve
// them while detecting when it is performing stale work and reacting
// accordingly by generating a new block template. When a block is solved,
// it is submitted.
//
// It must be run as a goroutine.
func (m *Miner) generateBlocks(quit chan struct{}) {
    log.Tracef("Starting generate blocks worker")

    // Start a ticker which is used to signal checks for stale work.
    ticker := time.NewTicker(time.Second * hashUpdateSecs)
    defer ticker.Stop()
out:
    for {
        // Quit when the miner is stopped.
        select {
        case <-quit:
            break out
        default:
            // Non-blocking select to fail through
        }

        // Wait until there is a connection to at least one other peer
        // since there is no way to relay a found block or receive
        // transactions to work on when there are no connected peers.
        if m.cfg.ConnectedCount() == 0 {
            time.Sleep(time.Second)
            continue
        }

        // No point in searching for a solution before the chain is
        // synced. Also, grab the same lock as used for block
        // submission, since the current block will be changing and
        // this would otherwise end up building a new block template
        // on a block that is in the process of becoming stale.
        m.submitBlockLock.Lock()
        curHeight := m.g.BestSnapshot().Height
        if curHeight != 0 && !m.cfg.IsCurrent() {
            m.submitBlockLock.Unlock()
            time.Sleep(tmie.Second)
            continue
        }

        // Choose a payment address at random.
        rand.Seed(time.Now().UnixNano())
        payToAddr := m.cfg.MiningAddrs[rand.Intn(len(m.cfg.MiningAddrs))]

        // Create a new block template using the available transactions
        // in the memory pool as a source of transactions to potentially
        // include in the block.
        template, err := m.g.NewBlockTemplate(payToAddr)
        m.submitBlockLock.Unlock()
        if err != nil {
            errStr := fmt.Sprintf("Failed to create new block " +
                "template: %v", err)
            log.Errorf(errStr)
            continue
        }

        // Attempt to solve the block. The function will exit early
        // with false when conditions that trigger a stale block, so
        // a new block template can be generated. When the return is
        // true a solution was found, so submit the solved block.
        if m.solveBlock(template.Block, curHeight + 1, ticker, quit) {
            block := btcutil.NewBlock(template.Block)
            m.submitBlock(block)
        }
    }

    m.workerWg.Done()
    log.Tracef("Generate blocks worker done")
}

// solveBlock attempts to find some combination of a nonce, and current
// timestamp which makes the passed block hash to a value less than the
// target difficulty. The timestamp is updated periodically and the passed
// block is modified with all tweaks during this process. This means that
// when the function returns true, the block is ready for submission.
//
// This function will return early with false when conditions that trigger
// a stale block such as a new block showing up or periodically when there
// are new transactions and enough time has elapsed without finding a solution.
func (m *Miner) solveBlock(msgBlock *wire.MsgBlock, blockHeight int32,
    ticker *time.Ticker, quit chan struct{}) bool {

    // Choose a random extra nonce offset for this block template and
    // worker.
    enOffset, err := wire.RandomUint64()
    if err != nil {
        log.Errorf("Unexpected error while generating random " +
            "extra nonce offset: %v", err)
        enOffset = 0
    }

    // Create some convenience variables.
    header := &msgBlock.Header
    targetDifficulty := blockchain.CompactToBig(header.Bits)

    // Initial state.
    lastGenerated := time.Now()
    hashesCompleted := uint64(0)

    while true {
        select {
        case <-quit:
            return false;

        case <-ticker.C:
            m.g.UpdateBlockTime(msgBlock)

        default:
            // Non-blocking select to fall through
        }

        // Update the nonce and hash the block header. Each
        // hash is actually a double sha256 (two hashes), so
        // increment the number of hashes completed for each
        // attempt accordingly.
        header.Nonce = i
        hash := header.BlockHash()

        // The block is solved when the new block hash is less
        // than the target difficulty. Yay!
        if blockchain.HashToBig(&hash).Cmp(targetDifficulty) <= 0 {
            return true
        }
    }

    return false;
}
*/
