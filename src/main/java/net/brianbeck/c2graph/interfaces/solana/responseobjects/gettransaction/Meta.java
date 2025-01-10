package net.brianbeck.c2graph.interfaces.solana.responseobjects.gettransaction;

import java.math.BigInteger;
import java.util.List;

import net.brianbeck.c2graph.interfaces.solana.responseobjects.general.InnerInstruction;
import net.brianbeck.c2graph.interfaces.solana.responseobjects.general.TokenBalance;

public class Meta {

    private Object err;
    private BigInteger fee;
    private List<InnerInstruction> innerInstructions;
    private List<BigInteger> postBalances;
    private List<TokenBalance> postTokenBalances;
    private List<BigInteger> preBalances;
    private List<TokenBalance> preTokenBalances;
    private List<Reward> rewards;
    private Status status;
    private BigInteger computeUnitsConsumed;
    private LoadedAddresses loadedAddresses;
    private List<String> logMessages;
    private String version;

    // Getters and Setters
    public Object getErr() {
        return err;
    }

    public void setErr(Object err) {
        this.err = err;
    }

    public BigInteger getFee() {
        return fee;
    }

    public void setFee(BigInteger fee) {
        this.fee = fee;
    }

    public List<InnerInstruction> getInnerInstructions() {
        return innerInstructions;
    }

    public void setInnerInstructions(List<InnerInstruction> innerInstructions) {
        this.innerInstructions = innerInstructions;
    }

    public List<BigInteger> getPostBalances() {
        return postBalances;
    }

    public void setPostBalances(List<BigInteger> postBalances) {
        this.postBalances = postBalances;
    }

    public List<TokenBalance> getPostTokenBalances() {
        return postTokenBalances;
    }

    public void setPostTokenBalances(List<TokenBalance> postTokenBalances) {
        this.postTokenBalances = postTokenBalances;
    }

    public List<BigInteger> getPreBalances() {
        return preBalances;
    }

    public void setPreBalances(List<BigInteger> preBalances) {
        this.preBalances = preBalances;
    }

    public List<TokenBalance> getPreTokenBalances() {
        return preTokenBalances;
    }

    public void setPreTokenBalances(List<TokenBalance> preTokenBalances) {
        this.preTokenBalances = preTokenBalances;
    }

    public List<Reward> getRewards() {
        return rewards;
    }

    public void setRewards(List<Reward> rewards) {
        this.rewards = rewards;
    }

    public Status getStatus() {
        return status;
    }

    public void setStatus(Status status) {
        this.status = status;
    }

    public BigInteger getComputeUnitsConsumed() {
        return computeUnitsConsumed;
    }

    public void setComputeUnitsConsumed(BigInteger computeUnitsConsumed) {
        this.computeUnitsConsumed = computeUnitsConsumed;
    }

    public LoadedAddresses getLoadedAddresses() {
        return loadedAddresses;
    }

    public void setLoadedAddresses(LoadedAddresses loadedAddresses) {
        this.loadedAddresses = loadedAddresses;
    }

    public List<String> getLogMessages() {
        return logMessages;
    }

    public void setLogMessages(List<String> logMessages) {
        this.logMessages = logMessages;
    }

    public String getVersion() {
        return version;
    }

    public void setVersion(String version) {
        this.version = version;
    }
}

