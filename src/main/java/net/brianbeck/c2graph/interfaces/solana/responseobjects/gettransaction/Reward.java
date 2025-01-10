package net.brianbeck.c2graph.interfaces.solana.responseobjects.gettransaction;

public class Reward {
    private String pubkey;
    private Long lamports;
    private Long postBalance;
    private String rewardType;
    private Integer commission;

    // Getters and Setters
    public String getPubkey() {
        return pubkey;
    }

    public void setPubkey(String pubkey) {
        this.pubkey = pubkey;
    }

    public Long getLamports() {
        return lamports;
    }

    public void setLamports(Long lamports) {
        this.lamports = lamports;
    }

    public Long getPostBalance() {
        return postBalance;
    }

    public void setPostBalance(Long postBalance) {
        this.postBalance = postBalance;
    }

    public String getRewardType() {
        return rewardType;
    }

    public void setRewardType(String rewardType) {
        this.rewardType = rewardType;
    }

    public Integer getCommission() {
        return commission;
    }

    public void setCommission(Integer commission) {
        this.commission = commission;
    }
}
