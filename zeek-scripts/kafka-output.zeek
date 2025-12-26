##! Zeek Kafka Integration
##! 将Zeek日志实时发送到Kafka

@load policy/frameworks/notice
@load base/frameworks/logging
@load packages/metron-bro-plugin-kafka/Kafka

module KafkaOutput;

export {
    ## Kafka broker地址
    const kafka_brokers = getenv("KAFKA_BROKERS") &redef;
    
    ## Topic前缀
    const topic_prefix = "zeek" &redef;
    
    ## 是否启用Kafka输出
    const enable_kafka = T &redef;
}

event zeek_init() &priority=-10 {
    if (!enable_kafka) {
        print "Kafka output disabled";
        return;
    }

    if (kafka_brokers == "") {
        Reporter::warning("KAFKA_BROKERS environment variable not set, using default");
        Kafka::kafka_conf = table(
            ["metadata.broker.list"] = "kafka:9092"
        );
    } else {
        Kafka::kafka_conf = table(
            ["metadata.broker.list"] = kafka_brokers
        );
    }

    Kafka::topic_name = "";
    Kafka::tag_json = T;
    Kafka::send_all_active_logs = T;
    
    local conn_filter: Log::Filter = [
        $name = "kafka-conn",
        $writer = Log::WRITER_KAFKAWRITER,
        $config = table(
            ["topic_name"] = fmt("%s-conn", topic_prefix)
        )
    ];
    Log::add_filter(Conn::LOG, conn_filter);
    
    local dns_filter: Log::Filter = [
        $name = "kafka-dns",
        $writer = Log::WRITER_KAFKAWRITER,
        $config = table(
            ["topic_name"] = fmt("%s-dns", topic_prefix)
        )
    ];
    Log::add_filter(DNS::LOG, dns_filter);
    
    local http_filter: Log::Filter = [
        $name = "kafka-http",
        $writer = Log::WRITER_KAFKAWRITER,
        $config = table(
            ["topic_name"] = fmt("%s-http", topic_prefix)
        )
    ];
    Log::add_filter(HTTP::LOG, http_filter);
    
    local ssl_filter: Log::Filter = [
        $name = "kafka-ssl",
        $writer = Log::WRITER_KAFKAWRITER,
        $config = table(
            ["topic_name"] = fmt("%s-ssl", topic_prefix)
        )
    ];
    Log::add_filter(SSL::LOG, ssl_filter);
    
    local files_filter: Log::Filter = [
        $name = "kafka-files",
        $writer = Log::WRITER_KAFKAWRITER,
        $config = table(
            ["topic_name"] = fmt("%s-files", topic_prefix)
        )
    ];
    Log::add_filter(Files::LOG, files_filter);
    
    local notice_filter: Log::Filter = [
        $name = "kafka-notice",
        $writer = Log::WRITER_KAFKAWRITER,
        $config = table(
            ["topic_name"] = fmt("%s-notice", topic_prefix)
        )
    ];
    Log::add_filter(Notice::LOG, notice_filter);

    print fmt("Kafka output enabled: brokers=%s, topic_prefix=%s", 
              kafka_brokers == "" ? "kafka:9092" : kafka_brokers, 
              topic_prefix);
}
