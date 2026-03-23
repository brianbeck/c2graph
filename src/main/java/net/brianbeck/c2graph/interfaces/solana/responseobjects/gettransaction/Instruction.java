package net.brianbeck.c2graph.interfaces.solana.responseobjects.gettransaction;

import java.util.List;
import java.math.BigInteger;

public class Instruction {
    private List<Integer> accounts;
    private String data;
    private int programIdIndex;
    private BigInteger stackHeight;

    // Getters and Setters
    public List<Integer> getAccounts() {
        return accounts;
    }

    public void setAccounts(List<Integer> accounts) {
        this.accounts = accounts;
    }

    public String getData() {
        return data;
    }

    public void setData(String data) {
        this.data = data;
    }

    public int getProgramIdIndex() {
        return programIdIndex;
    }

    public void setProgramIdIndex(int programIdIndex) {
        this.programIdIndex = programIdIndex;
    }

    public BigInteger getStackHeight() {
        return stackHeight;
    }

    public void setStackHeight(BigInteger stackHeight) {
        this.stackHeight = stackHeight;
    }
}