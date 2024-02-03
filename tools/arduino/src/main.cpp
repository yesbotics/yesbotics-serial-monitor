#include <Arduino.h>


int index = 0;

void setup() {
    // Starte die serielle Kommunikation mit 9600 Baud
    Serial.begin(57600);
}

void loop() {

    // Entscheide, welcher Text basierend auf der Variablen textNummer ausgegeben wird
    switch(index) {
        case 0:
            Serial.println("Textmessage");
            break;
        case 1:
            Serial.println("Next has unprintable characters");
            break;
        case 2:
            Serial.write('a');
            Serial.write(0x00);
            Serial.write(0x04);
            Serial.write(0x0a);
            break;
    }

    index = (index + 1) % 3;
    delay(1000);
}
