package net.brianbeck.c2graph.interfaces.solana.responseobjects.gettokenaccountsbyowner;

import javax.xml.crypto.Data;

public class Account {

        private Data data;
    private boolean executable;
    private long lamports;
    private String owner;
    private long rentEpoch;
    private int space;

    // Getters and Setters
    public Data getData() {
        return data;
    }

    public void setData(Data data) {
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

    public long getRentEpoch() {
        return rentEpoch;
    }

    public void setRentEpoch(long rentEpoch) {
        this.rentEpoch = rentEpoch;
    }

    public int getSpace() {
        return space;
    }

    public void setSpace(int space) {
        this.space = space;
    }
}
