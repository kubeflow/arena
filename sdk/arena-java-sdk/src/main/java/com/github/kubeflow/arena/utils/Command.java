package com.github.kubeflow.arena.utils;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.nio.charset.Charset;
import java.util.ArrayList;
import java.util.List;

import com.github.kubeflow.arena.exceptions.ExitCodeException;
import org.apache.commons.lang3.StringUtils;

public class Command {

    private String[] mCommand;

    private Command(String[] execString) {
        mCommand = execString.clone();
    }

    /**
     * Runs a command and returns its stdout on success.
     *
     * @return the output
     * @throws ExitCodeException if the command returns a non-zero exit code
     */
    private String run() throws ExitCodeException, IOException {
        Process process = new ProcessBuilder(mCommand).redirectErrorStream(true).start();

        BufferedReader inReader =
                new BufferedReader(new InputStreamReader(process.getInputStream(),
                        Charset.defaultCharset()));

        try {
            // read the output of the command
            StringBuilder output = new StringBuilder();
            String line = inReader.readLine();
            while (line != null) {
                output.append(line);
                output.append("\n");
                line = inReader.readLine();
            }
            // wait for the process to finish and check the exit code
            int exitCode = process.waitFor();
            if (exitCode != 0) {
                throw new ExitCodeException(exitCode, output.toString());
            }
            return output.toString();
        } catch (InterruptedException e) {
            e.printStackTrace();
            throw new IOException(e.getMessage());
        } finally {
            // close the input stream
            try {
                // JDK 7 tries to automatically drain the input streams for us
                // when the process exits, but since close is not synchronized,
                // it creates a race if we close the stream first and the same
                // fd is recycled. the stream draining thread will attempt to
                // drain that fd!! it may block, OOM, or cause bizarre behavior
                // see: https://bugs.openjdk.java.net/browse/JDK-8024521
                // issue is fixed in build 7u60
                InputStream stdout = process.getInputStream();
                synchronized (stdout) {
                    inReader.close();
                }
            } catch (IOException e) {
                System.out.printf("Error while closing the input stream", e);
            }
            process.destroy();
        }
    }

    /**
     * Static method to execute a shell command.
     *
     * @param cmd shell command to execute
     * @return the output of the executed command
     */
    public static String execCommand(String... cmd) throws IOException {
        String cmdString = StringUtils.join(cmd, " ");
        System.out.printf("exec command: [%s]\n",cmdString);
        return String.format("exec command: [%s]\n", cmdString) + new Command(newCommandBuild(cmd)).run();
    }

    private static String[] newCommandBuild(String... cmd) {
        List<String> newCommand = new ArrayList<String>();
        for (String s : cmd) {
            if (s.contains("--env") || s.contains("--annotation") || s.contains("--label")){
                newCommand.add(s.replaceAll("\"",""));
            } else {
                newCommand.add(s);
            }
        }
        return newCommand.toArray(new String[newCommand.size()]);
    }

}