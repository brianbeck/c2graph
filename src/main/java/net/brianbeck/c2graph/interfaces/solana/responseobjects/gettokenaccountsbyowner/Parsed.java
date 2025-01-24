package net.brianbeck.c2graph.interfaces.solana.responseobjects.gettokenaccountsbyowner;

public class Parsed {

    private Info info;
    private String type;

    // Getters and Setters
    public Info getInfo() {
        return info;
    }

    public void setInfo(Info info) {
        this.info = info;
    }

    public String getType() {
        return type;
    }

    public void setType(String type) {
        this.type = type;
    }
}
