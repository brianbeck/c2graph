package net.brianbeck.c2graph.interfaces;

import net.brianbeck.c2graph.interfaces.solana.SolanaInterface;

public class CryptoInterfaceFactory {

    public enum BlockchainType {
        SOLANA,
        ETHEREUM
    }

    public static CryptoInterface getCryptoInterface(BlockchainType type) {
        switch (type) {
            case SOLANA:
                return new SolanaInterface();
            case ETHEREUM:
                throw new IllegalArgumentException("Unknown BlockchainType: " + type);
            default:
                throw new IllegalArgumentException("Unknown BlockchainType: " + type);
        }
    }
}