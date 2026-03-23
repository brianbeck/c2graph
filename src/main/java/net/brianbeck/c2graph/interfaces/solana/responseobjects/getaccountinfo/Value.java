package net.brianbeck.c2graph.interfaces.solana.responseobjects.getaccountinfo;

import java.util.List;
import java.math.BigInteger;

public class Value {
    private List<String> data;
    private boolean executable;
    private long lamports;
    private String owner;
    private BigInteger rentEpoch;
    private int space;

    // Getters and Setters
    public List<String> getData() {
        return data;
    }

    public void setData(List<String> data) {
        this.data = data;
    }

    public boolean isExecutable() {
        return executable;
    }

    public void setExecutable(boolean executable) {
        this.executable = executable;
    }

    public long getLamports() {
        return lamports;
    }

    public void setLamports(long lamports) {
        this.lamports = lamports;
    }

    public String getOwner() {
        return owner;
    }

    public void setOwner(String owner) {
        this.owner = owner;
    }

    public BigInteger getRentEpoch() {
        return rentEpoch;
    }

    public void setRentEpoch(BigInteger rentEpoch) {
        this.rentEpoch = rentEpoch;
    }

    public int getSpace() {
        return space;
    }

    public void setSpace(int space) {
        this.space = space;
    }
}
