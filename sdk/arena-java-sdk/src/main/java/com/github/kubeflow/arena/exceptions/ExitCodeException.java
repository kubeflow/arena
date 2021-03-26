package com.github.kubeflow.arena.exceptions;

import java.io.IOException;

/**
 * This is an IOException with exit code added.
 */
public  class ExitCodeException extends IOException {


    private final int mExitCode;

    /**
     * Constructs an ExitCodeException.
     *
     * @param exitCode the exit code returns by shell
     * @param message the exception message
     */
    public ExitCodeException(int exitCode, String message) {
        super(message);
        mExitCode = exitCode;
    }

    /**
     * Gets the exit code.
     *
     * @return the exit code
     */
    public int getExitCode() {
        return mExitCode;
    }

    @Override
    public String toString() {
        final StringBuilder sb = new StringBuilder("ExitCodeException ");
        sb.append("exitCode=").append(mExitCode).append(": ");
        sb.append(super.getMessage());
        return sb.toString();
    }
}