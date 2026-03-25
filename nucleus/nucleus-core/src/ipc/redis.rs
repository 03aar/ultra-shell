use redis::Commands;
use std::sync::mpsc;
use std::thread;

pub struct RedisPublisher {
    client: redis::Client,
}

impl RedisPublisher {
    pub fn new(redis_url: &str) -> Result<Self, redis::RedisError> {
        let client = redis::Client::open(redis_url)?;
        // Test connection
        let mut conn = client.get_connection()?;
        let _: () = redis::cmd("PING").query(&mut conn)?;
        Ok(RedisPublisher { client })
    }

    pub fn new_with_fallback(redis_url: &str) -> Self {
        match Self::new(redis_url) {
            Ok(publisher) => {
                log::info!("Connected to Redis at {}", redis_url);
                publisher
            }
            Err(e) => {
                log::warn!("Redis unavailable ({}), using no-op publisher", e);
                RedisPublisher {
                    client: redis::Client::open("redis://127.0.0.1:6379").unwrap(),
                }
            }
        }
    }

    pub fn publish(&self, channel: &str, message: &str) -> bool {
        match self.client.get_connection() {
            Ok(mut conn) => match conn.publish::<_, _, i32>(channel, message) {
                Ok(_) => true,
                Err(e) => {
                    log::debug!("Redis publish failed: {}", e);
                    false
                }
            },
            Err(e) => {
                log::debug!("Redis connection failed: {}", e);
                false
            }
        }
    }

    pub fn publish_execution(&self, execution_json: &str) -> bool {
        self.publish("nucleus:executions", execution_json)
    }

    pub fn publish_rollback(&self, rollback_json: &str) -> bool {
        self.publish("nucleus:rollbacks", rollback_json)
    }

    pub fn publish_session_event(&self, event_json: &str) -> bool {
        self.publish("nucleus:sessions", event_json)
    }
}

pub struct RedisSubscriber {
    client: redis::Client,
}

impl RedisSubscriber {
    pub fn new(redis_url: &str) -> Result<Self, redis::RedisError> {
        let client = redis::Client::open(redis_url)?;
        Ok(RedisSubscriber { client })
    }

    pub fn subscribe(
        &self,
        channels: &[&str],
        callback: impl Fn(&str, &str) + Send + 'static,
    ) -> Result<(), redis::RedisError> {
        let client = self.client.clone();
        let channels: Vec<String> = channels.iter().map(|s| s.to_string()).collect();

        thread::spawn(move || {
            if let Ok(mut conn) = client.get_connection() {
                let mut pubsub = conn.as_pubsub();
                for channel in &channels {
                    let _ = pubsub.subscribe(channel);
                }
                loop {
                    if let Ok(msg) = pubsub.get_message() {
                        let channel: String = msg.get_channel_name().to_string();
                        if let Ok(payload) = msg.get_payload::<String>() {
                            callback(&channel, &payload);
                        }
                    }
                }
            }
        });

        Ok(())
    }
}
