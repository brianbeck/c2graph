package net.brianbeck.c2graph.interfaces.solana.responseobjects.gettokenaccountsbyowner;

public class TokenAmount {
    
    private String amount;
    private int decimals;
    private Double uiAmount;
    private String uiAmountString;

    public String getAmount() {
        return amount;
    }

    public void setAmount(String amount) {
        this.amount = amount;
    }

    public int getDecimals() {
        return decimals;
    }

    public void setDecimals(int decimals) {
        this.decimals = decimals;
    }

    public Double getUiAmount() {
        return uiAmount;
    }

    public void setUiAmount(Double uiAmount) {
        this.uiAmount = uiAmount;
    }

    public String getUiAmountString() {
        return uiAmountString;
    }

    public void setUiAmountString(String uiAmountString) {
        this.uiAmountString = uiAmountString;
    }
}