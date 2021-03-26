package com.github.kubeflow.arena.model.fields;

import com.github.kubeflow.arena.enums.ArenaErrorEnum;
import com.github.kubeflow.arena.exceptions.ArenaException;

import java.util.ArrayList;

public class StringListField implements Field {
    private String flag;
    private ArrayList<String> values;

    public StringListField(String flag,ArrayList<String> values) {
        this.flag = flag;
        this.values = values;
    }

    public void validate() throws ArenaException {
        if (this.values == null || this.values.size() == 0) {
            throw  new ArenaException(ArenaErrorEnum.VALIDATE_ARGS,"value of option " + this.flag + " is null");
        }
    }

    public ArrayList<String> options() {
        ArrayList<String> options = new ArrayList<String>();
        for (int i = 0;i < this.values.size();i++) {
            options.add(this.flag +"="+ this.values.get(i));
        }
        return options;
    }

}
