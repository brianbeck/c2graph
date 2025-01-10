package net.brianbeck.c2graph.interfaces.solana.responseobjects.general;

import java.util.List;

public class InnerInstruction {
    private int index;
    private List<InstructionDetail> instructions;

    // Getters and Setters
    public int getIndex() {
        return index;
    }

    public void setIndex(int index) {
        this.index = index;
    }

    public List<InstructionDetail> getInstructions() {
        return instructions;
    }

    public void setInstructions(List<InstructionDetail> instructions) {
        this.instructions = instructions;
    }
}