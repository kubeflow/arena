package com.github.kubeflow.arena.model.fields;

import com.github.kubeflow.arena.enums.ArenaErrorEnum;
import com.github.kubeflow.arena.exceptions.ArenaException;

import java.util.ArrayList;
import java.util.Map;

public class StringMapField implements Field {
    private String flag;
    private Map<String, String> values;
    private String joinFlag = "=";

    public StringMapField(String flag,Map<String, String> values) {
        this.flag = flag;
        this.values = values;
    }

    public StringMapField(String flag,Map<String, String> values,String joinFlag) {
        this.flag = flag;
        this.values = values;
        this.joinFlag = joinFlag;
    }

    public void validate() throws ArenaException {
        if (this.values == null || this.values.size() == 0) {
            throw  new ArenaException(ArenaErrorEnum.VALIDATE_ARGS,"value of option " + this.flag + " is null");
        }
    }

    public ArrayList<String> options() {
        ArrayList<String> options = new ArrayList<String>();
        for (Map.Entry<String, String> entry : this.values.entrySet()) {
            options.add(this.flag +"="+ entry.getKey() + this.joinFlag + entry.getValue());
        }
        return options;
    }
}
