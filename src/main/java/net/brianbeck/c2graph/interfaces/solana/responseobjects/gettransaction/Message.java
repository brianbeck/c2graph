package net.brianbeck.c2graph.interfaces.solana.responseobjects.gettransaction;

import java.util.List;

public class Message {
    private List<String> accountKeys;
    private Header header;
    private List<Instruction> instructions;
    private String recentBlockhash;

    // Getters and Setters
    public List<String> getAccountKeys() {
        return accountKeys;
    }

    public void setAccountKeys(List<String> accountKeys) {
        this.accountKeys = accountKeys;
    }

    public Header getHeader() {
        return header;
    }

    public void setHeader(Header header) {
        this.header = header;
    }

    public List<Instruction> getInstructions() {
        return instructions;
    }

    public void setInstructions(List<Instruction> instructions) {
        this.instructions = instructions;
    }

    public String getRecentBlockhash() {
        return recentBlockhash;
    }

    public void setRecentBlockhash(String recentBlockhash) {
        this.recentBlockhash = recentBlockhash;
    }
}

