package net.brianbeck.c2graph.interfaces.solana.responseobjects.gettokenaccountsbyowner;

public class Info {

    private boolean isNative;
    private String mint;
    private String owner;
    private String state;
    private TokenAmount tokenAmount;

    // Getters and Setters
    public boolean isNative() {
        return isNative;
    }

    public void setNative(boolean isNative) {
        this.isNative = isNative;
    }

    public String getMint() {
        return mint;
    }

    public void setMint(String mint) {
        this.mint = mint;
    }

    public String getOwner() {
        return owner;
    }

    public void setOwner(String owner) {
        this.owner = owner;
    }

    public String getState() {
        return state;
    }

    public void setState(String state) {
        this.state = state;
    }

    public TokenAmount getTokenAmount() {
        return tokenAmount;
    }

    public void setTokenAmount(TokenAmount tokenAmount) {
        this.tokenAmount = tokenAmount;
    }

}
