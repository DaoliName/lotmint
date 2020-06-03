/* Bitcoin-NG/ByzCoin 提出的KeyBlock实现 */

/*
type KeyBlock struct {
    //! pointer to the hash of the block, if any. Memory is owned by this CBlockIndex
    const uint256* phashBlock;

    //! pointer to the index of the predecessor of this block
    CBlockIndex* pprev;

    //! height of the entry in the chain. The genesis block has height 0
    int nHeight;

    uint32_t nBits;
    uint32_t nNonce;

    ...
}
*/
