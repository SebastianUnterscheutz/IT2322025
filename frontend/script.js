document.getElementById('offerForm').addEventListener('submit', function (event) {
    event.preventDefault();
    var formData = new FormData(event.target);

    // Behalte die JSON-Reihenfolge durch Definition der Objektdaten in der gewünschten Reihenfolge
    var data = {
        name: formData.get('name'),
        first_name: formData.get('vorname'),
        email: formData.get('email'),
        class: formData.get('klasse'),
        phone_number: formData.get('handy'),
        valid_from: formData.get('gueltig_von'),
        valid_until: formData.get('gueltig_bis'),
        additional_information: formData.get('info'),
        other: "Zusätzliche Angaben", // Optional, falls benötigt
        offer_locations: [] // Orte werden hier eingefügt
    };

    // Orte sammeln
    var locations = document.querySelectorAll('[name^="ort"], [name^="plz"]');
    for (var i = 0; i < locations.length; i += 2) {
        data.offer_locations.push({
            plz: locations[i + 1].value,
            city: locations[i].value,
            street: formData.get('strasse') || "", // Sicherstellen, dass Straßeninformationen leer sein können
            house_number: formData.get('hausnummer') || "" // Sicherstellen, dass Hausnummern leer sein können
        });
    }

    // Daten an das Backend senden
    fetch('/api/create/offer', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(data) // JSON in der exakten Reihenfolge wie definiert
    }).then(response => {
        if (response.ok) {
            alert('Angebot erfolgreich erstellt.');
        } else {
            alert('Fehler beim Erstellen des Angebots.');
        }
    }).catch(error => {
        alert('Verbindungsfehler: ' + error.message);
    });
});