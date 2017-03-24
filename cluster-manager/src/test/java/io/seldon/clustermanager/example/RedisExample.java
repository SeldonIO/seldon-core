package io.seldon.clustermanager.example;

import org.springframework.boot.builder.SpringApplicationBuilder;
import org.springframework.context.ConfigurableApplicationContext;
import org.springframework.data.redis.connection.RedisConnectionFactory;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.data.redis.core.ValueOperations;

import io.seldon.clustermanager.config.RedisConfig;

public class RedisExample {
    public static void main(String[] args) {
        ConfigurableApplicationContext ctx = new SpringApplicationBuilder(RedisConfig.class).web(false).run(args);

        RedisConnectionFactory rcf = ctx.getBean(RedisConnectionFactory.class);

        StringRedisTemplate redisTemplate = ctx.getBean(StringRedisTemplate.class);
        ValueOperations<String, String> values = redisTemplate.opsForValue();

        { // set a key
            values.set("mykey", "myvalue");
        }
        { // get a value for key

            String v = values.get("mykey");
            System.out.println(v);
        }

        ctx.close();
    }

}
