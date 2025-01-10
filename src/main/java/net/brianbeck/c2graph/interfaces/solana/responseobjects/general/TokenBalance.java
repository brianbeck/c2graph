package net.brianbeck.c2graph.interfaces.solana.responseobjects.general;

public class TokenBalance {
    
    private int accountIndex;
    private String mint;
    private String owner; // Can be null or undefined
    private String programId; // Can be null or undefined
    private UiTokenAmount uiTokenAmount;

    // Getters and Setters
    public int getAccountIndex() {
        return accountIndex;
    }

    public void setAccountIndex(int accountIndex) {
        this.accountIndex = accountIndex;
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

    public String getProgramId() {
        return programId;
    }

    public void setProgramId(String programId) {
        this.programId = programId;
    }

    public UiTokenAmount getUiTokenAmount() {
        return uiTokenAmount;
    }

    public void setUiTokenAmount(UiTokenAmount uiTokenAmount) {
        this.uiTokenAmount = uiTokenAmount;
    }
}
