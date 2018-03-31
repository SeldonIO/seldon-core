package io.seldon.wrapper.api.router;

import org.springframework.boot.autoconfigure.condition.ConditionalOnExpression;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.RestController;

@RestController
@ConditionalOnExpression("${seldon.api.route.enabled:false}")
public class RouterRestController {

	@RequestMapping(value = "/route", method = RequestMethod.GET)
    String ping() {
        return "ROUTE";
    }
	
	
}