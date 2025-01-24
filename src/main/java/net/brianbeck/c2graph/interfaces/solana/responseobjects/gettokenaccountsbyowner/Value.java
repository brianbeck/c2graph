package net.brianbeck.c2graph.interfaces.solana.responseobjects.gettokenaccountsbyowner;

import java.util.List;

public class Value {

    private List<Account> account;
    private String pubkey;

    public List<Account> getAccount() {
        return account;
    }

    public void setAccount(List<Account> account) {
        this.account = account;
    }

    public String getPubkey() {
        return pubkey;
    }

    public void setPubkey(String pubkey) {
        this.pubkey = pubkey;
    }
}
