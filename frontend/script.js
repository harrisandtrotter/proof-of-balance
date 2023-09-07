document.addEventListener('DOMContentLoaded', function () {
    const form = document.getElementById('balanceForm');
    const resultDiv = document.getElementById('result');

    form.addEventListener('submit', async function (e) {
        e.preventDefault();

        const address = form.elements.address.value;
        const chain = form.elements.chain.value;
        const date = form.elements.date.value;
        const timestamp = form.elements.timestamp.value;

        const response = await fetch('http://localhost:8000/balances', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                address: address,
                chain: chain,
                date: date,
                timestamp: timestamp,
            }),
        });

        if (response.ok) {
            const data = await response.json();

            resultDiv.innerHTML = JSON.stringify(data, null, 2);
        } else {
            resultDiv.innerHTML = 'Error fetching data from the server. '
        }
    });
});