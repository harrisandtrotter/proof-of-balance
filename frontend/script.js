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
            // resultDiv.innerHTML = ''; // Clear any previous content

            // Create a table element
            const table = document.createElement('table');
            const cols = Object.keys(data[0]);
            const thead = document.createElement('thead');
            const tr = document.createElement('tr');

            cols.forEach((item) => {
                const th = document.createElement('th');
                th.innerText = item;
                tr.appendChild(th);
            });

            thead.appendChild(tr);
            table.append(tr);

            data.forEach((item) => {
                const tr = document.createElement('tr');
                const vals = Object.values(item);

                vals.forEach((elem) => {
                    const td = document.createElement('td');
                    td.innerText = elem;
                    tr.appendChild(td);
                });

                table.appendChild(tr);
            });

            resultDiv.appendChild(table);
        } else {
            resultDiv.innerHTML = 'Error fetching data from the server. ';
        }
    });
});
