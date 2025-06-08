Para executar o código siga os comandos abaixo:

sudo docker build -t load-tester .

sudo docker run --rm load-tester --url=http://google.com --requests=100 --concurrency=10

Após isso deverá ter o retorno esperado para o execicio conforme o print abaixo.

![image](https://github.com/user-attachments/assets/c348046a-3637-423f-9bc0-84bf258a5cd6)
