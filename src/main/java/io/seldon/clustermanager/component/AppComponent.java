package io.seldon.clustermanager.component;

public interface AppComponent {

    public void init() throws Exception;

    public void cleanup() throws Exception;
}
