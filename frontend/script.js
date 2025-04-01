const apiEndpoint = '/api/create/offer'; // Replace with your API endpoint
async function handleOfferFormSubmit(event) {
    event.preventDefault();

    const formData = new FormData(event.target);
    const formObject = Object.fromEntries(formData.entries());
    try {
        const response = await fetch(apiEndpoint, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(formObject)
        });
        if (response.ok) {
            const result = await response.json();
            alert('Mitfahrgelegenheit wurde erfolgreich eingereicht.');
            window.location.href = 'index.html';
        } else {
            alert('Fehler beim Einreichen der Mitfahrgelegenheit. Bitte versuchen Sie es erneut.');
        }
    } catch (error) {
        console.error('Error:', error);
        alert('Ein Netzwerkfehler ist aufgetreten. Bitte versuchen Sie es später erneut.');
    }
}


document.addEventListener('DOMContentLoaded', function() {
    const offerForm = document.getElementById('offerForm');
    if (offerForm) {
        offerForm.addEventListener('submit', handleOfferFormSubmit);
    }
});