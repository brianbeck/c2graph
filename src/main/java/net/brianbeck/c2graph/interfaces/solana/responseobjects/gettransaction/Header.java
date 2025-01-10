package net.brianbeck.c2graph.interfaces.solana.responseobjects.gettransaction;

public class Header {
    private int numReadonlySignedAccounts;
    private int numReadonlyUnsignedAccounts;
    private int numRequiredSignatures;

    // Getters and Setters
    public int getNumReadonlySignedAccounts() {
        return numReadonlySignedAccounts;
    }

    public void setNumReadonlySignedAccounts(int numReadonlySignedAccounts) {
        this.numReadonlySignedAccounts = numReadonlySignedAccounts;
    }

    public int getNumReadonlyUnsignedAccounts() {
        return numReadonlyUnsignedAccounts;
    }

    public void setNumReadonlyUnsignedAccounts(int numReadonlyUnsignedAccounts) {
        this.numReadonlyUnsignedAccounts = numReadonlyUnsignedAccounts;
    }

    public int getNumRequiredSignatures() {
        return numRequiredSignatures;
    }

    public void setNumRequiredSignatures(int numRequiredSignatures) {
        this.numRequiredSignatures = numRequiredSignatures;
    }
}