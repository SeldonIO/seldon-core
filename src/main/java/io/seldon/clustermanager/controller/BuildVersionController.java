package io.seldon.clustermanager.controller;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class BuildVersionController {

    @Value("${build.version}")
    private String buildVersion;

    @RequestMapping(value = "/version", method = RequestMethod.GET, produces = "text/plain; charset=utf-8")
    public String get_buildVersion() {
        return buildVersion;
    }

}
