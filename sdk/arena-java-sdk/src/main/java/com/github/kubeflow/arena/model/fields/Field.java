package com.github.kubeflow.arena.model.fields;
import com.github.kubeflow.arena.exceptions.ArenaException;

import java.util.*;

public interface Field {

    void validate() throws ArenaException;

    ArrayList<String> options();
}
