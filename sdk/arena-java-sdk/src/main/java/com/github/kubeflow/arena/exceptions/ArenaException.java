package com.github.kubeflow.arena.exceptions;

import com.github.kubeflow.arena.enums.ArenaErrorEnum;

public class ArenaException extends Exception {

    private ArenaErrorEnum exceptionEnum;

    private String message;

    public ArenaException(ArenaErrorEnum exceptionEnum){
        super(exceptionEnum.message);
        this.exceptionEnum = exceptionEnum;
        this.message = exceptionEnum.message;
    }

    public ArenaException(ArenaErrorEnum exceptionEnum,String message){
        super(message);
        this.exceptionEnum = exceptionEnum;
        this.message = message;
    }

    public ArenaException(ArenaErrorEnum exceptionEnum,Throwable throwable){
        super(exceptionEnum.message,throwable);
        this.exceptionEnum = exceptionEnum;
        this.message = exceptionEnum.message;
    }

    public ArenaException(ArenaErrorEnum exceptionEnum,String message,Throwable throwable){
        super(message,throwable);
        this.exceptionEnum = exceptionEnum;
        this.message = message;
    }

    public ArenaErrorEnum errorType() {
        return this.exceptionEnum;
    }

    public String message() {
        return this.message;
    }

    @Override
    public String toString() {
        final StringBuilder sb = new StringBuilder("ArenaException  ");
        sb.append("ErrType: ").append(this.exceptionEnum.name()).append(", ");
        sb.append("ErrMessage: ").append(super.getMessage());
        return sb.toString();
    }

    @Override
    public void printStackTrace(){
        System.out.println(this.toString());
    }
}
