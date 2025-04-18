document.getElementById('offerForm').addEventListener('submit',getData)

function getData(event) {
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
        offer_locations: [{
            plz: formData.get('plz'),
            city: formData.get('ort'),
            street: formData.get('strasse'),
            house_number: formData.get('hausnummer')
        }] // Orte werden hier eingefügt
    };

    // Orte sammeln
    var filteredLocations = locations.filter(item => item !== null);
    for (var i = 0; i < filteredLocations.length; i++) {
        data.offer_locations.push(filteredLocations[i]);
    }

    document.getElementById('anbieten').disabled = true;

    // Daten an das Backend senden
    fetch('/api/create/offer', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(data) // JSON in der exakten Reihenfolge wie definiert
    }).then(response => response.json())
      .then(data => {
        if (data.status == 'success') {
            alert('Angebot erfolgreich erstellt.');
            window.location.href = "/";
        } else {
            alert(`
                Fehler beim Erstellen des Angebots:

                ${data.message}`);
            };
            document.getElementById('anbieten').disabled = false
    }).catch(error => {
        alert('Verbindungsfehler: ' + error.message);
    });
};

document.addEventListener('DOMContentLoaded', function() {
    const offerForm = document.getElementById('offerForm');
    if (offerForm) {
        // offerForm.addEventListener('submit', handleOfferFormSubmit);
        var date = new Date();
        var year = date.getFullYear();
        var month = String(date.getMonth() + 1).padStart(2, '0');
        var day = String(date.getDate()).padStart(2, '0');
        document.getElementById('gueltig_von').value = `${year}-${month}-${day}`;
        console.log(`${year}-${month}-${day}`);
        date.setMonth(date.getMonth() + 6);
        var year = date.getFullYear();
        var month = String(date.getMonth() + 1).padStart(2, '0');
        var day = String(date.getDate()).padStart(2, '0');
        document.getElementById('gueltig_bis').value = `${year}-${month}-${day}`;
    }
});

var map;
var locations = [];
document.markers = [];

// handle interactive map to add new locations on the way
function loadMapInForm() {
    map = L.map('map').setView([50.527724, 12.402964], 13);
    L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
        maxZoom: 19,
        attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
    }).addTo(map);
    var marker = L.marker([50.527724, 12.402964]).addTo(map);
    marker.bindPopup("<b>BSZ Rodewisch</b>")
    map.on('click', onMapClick)
}
function onMapClick(e) {
    // Marker auf Karte anzeigen und Button zum Entfernen hinzufügen
    var marker = L.marker(e.latlng).addTo(map);
    document.markers.push(marker)
    marker.bindPopup(`
        <button type="button" onclick="removeMarker(${document.markers.length-1});">Entfernen</button>
    `).openPopup();
    // Ort vorbereiten zum Anschicken
    locations.push({
        latitude: e.latlng.lat,
        longitude: e.latlng.lng
    })
}

function removeMarker(index) {
    // Marker von Karte und Orte Array entfernen
    map.removeLayer(document.markers[index]);
    locations[index] = null;
}
