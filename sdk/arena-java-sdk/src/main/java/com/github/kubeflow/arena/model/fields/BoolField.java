package com.github.kubeflow.arena.model.fields;

import com.github.kubeflow.arena.exceptions.ArenaException;

import java.util.ArrayList;

public class BoolField implements Field {
    private String flag;

    public BoolField(String flag) {
        this.flag = flag;
    }

    public void validate() throws ArenaException {

    }

    public ArrayList<String> options() {
        ArrayList<String> options = new ArrayList<String>();
        options.add(this.flag);
        return options;
    }
}
