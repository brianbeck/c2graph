package net.brianbeck.c2graph.interfaces.solana.responseobjects.general;

import java.util.List;

public class InstructionDetail {
    private int programIdIndex;
    private List<Integer> accounts;
    private String data;

    // Getters and Setters
    public int getProgramIdIndex() {
        return programIdIndex;
    }

    public void setProgramIdIndex(int programIdIndex) {
        this.programIdIndex = programIdIndex;
    }

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
}