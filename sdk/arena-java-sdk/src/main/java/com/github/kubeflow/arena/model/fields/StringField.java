package com.github.kubeflow.arena.model.fields;

import com.github.kubeflow.arena.exceptions.ArenaException;
import com.github.kubeflow.arena.enums.ArenaErrorEnum;

import java.util.*;

public class StringField implements Field {
    private String flag;
    private String value;

    public StringField(String flag,String value) {
        this.flag = flag;
        this.value = value;
    }

    public void validate() throws ArenaException {
        if (this.value == null || this.value.length() == 0) {
            throw new ArenaException(ArenaErrorEnum.VALIDATE_ARGS,"value of option " + this.flag + " is null");
        }
    }

    public ArrayList<String> options() {
        ArrayList<String> options = new ArrayList<String>();
        options.add(this.flag +"="+ this.value);
        return options;
    }

}
