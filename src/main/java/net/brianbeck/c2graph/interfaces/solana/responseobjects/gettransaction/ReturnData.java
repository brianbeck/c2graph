package net.brianbeck.c2graph.interfaces.solana.responseobjects.gettransaction;

import java.util.List;

public class ReturnData {
    private String programId;
    private List<String> data;

    // Getters and Setters
    public String getProgramId() {
        return programId;
    }

    public void setProgramId(String programId) {
        this.programId = programId;
    }

    public List<String> getData() {
        return data;
    }

    public void setData(List<String> data) {
        this.data = data;
    }
}