package net.brianbeck.c2graph.interfaces.solana.responseobjects.gettokenaccountsbyowner;

public class Data {

    private Parsed parsed;
    private String program;
    private int space;

    public Parsed getParsed() {
        return parsed;
    }

    public void setParsed(Parsed parsed) {
        this.parsed = parsed;
    }

    public String getProgram() {
        return program;
    }

    public void setProgram(String program) {
        this.program = program;
    }

    public int getSpace() {
        return space;
    }

    public void setSpace(int space) {
        this.space = space;
    }
}
