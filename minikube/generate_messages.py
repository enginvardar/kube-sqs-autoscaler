import os
import sys

def generateMessages(queue_name, port, amount):
    for i in range(int(amount)):
        a = "aws sqs --region eu-central-1 --endpoint-url http://127.0.0.1:"+port+" send-message --queue-url http://localhost:9324/queue/"+queue_name+" --message-body 'come random message' &"
        os.system(a)

if __name__ == "__main__":
    generateMessages(sys.argv[1], sys.argv[2], sys.argv[3])