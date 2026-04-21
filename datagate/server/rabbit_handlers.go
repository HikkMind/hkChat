package server

func (server *DatabaseServer) handleRabbitConnection() {
	if server.rabbitChannel == nil {
		server.logger.Fatal("rabbitChannel is nil, cannot start consumer")
		return
	}

	if err := server.rabbitChannel.Qos(10, 0, false); err != nil {
		server.logger.Printf("failed to set QoS: %v", err)
		return
	}

	deliveryChan, err := server.rabbitChannel.Consume(
		"datagate.queue",    // имя очереди (должна быть создана в rabbitInit)
		"datagate-consumer", // consumer tag
		false,               // auto-ack (false - ручное подтверждение)
		false,               // exclusive
		false,               // no-local
		false,               // no-wait
		nil,                 // args
	)
	if err != nil {
		server.logger.Fatalf("failed to register consumer: %v", err)
		return
	}

	server.logger.Print("RabbitMQ consumer started, waiting for messages...")

	go func() {
		for message := range deliveryChan {
			server.processMessage(message)
		}
	}()
	go server.handleDQLMessages()
}

func (server *DatabaseServer) handleDQLMessages() {
	deliveryChan, err := server.rabbitChannel.Consume("datagate.dlq", "dlq-consumer", false, false, false, false, nil)
	if err != nil {
		server.logger.Fatal(err)
	}
	for message := range deliveryChan {
		server.logger.Printf("DLQ message: %s", message.Body)
		message.Ack(false)
	}
}
