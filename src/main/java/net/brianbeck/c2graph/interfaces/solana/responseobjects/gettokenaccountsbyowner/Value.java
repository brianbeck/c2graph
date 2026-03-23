package net.brianbeck.c2graph.interfaces.solana.responseobjects.gettokenaccountsbyowner;

public class Value {

    private Account account;
    private String pubkey;

    public Account getAccount() {
        return account;
    }

    public void setAccount(Account account) {
        this.account = account;
    }

    public String getPubkey() {
        return pubkey;
    }

    public void setPubkey(String pubkey) {
        this.pubkey = pubkey;
    }
}
