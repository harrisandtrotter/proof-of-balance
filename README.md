# proof-of-balance
A tool to check the balance of any address at the specified time (block number) for most EVM-compatible chains.
Useful for audit testing and gains assurance over the following assertions: completeness and existence. Further functionality will be added in future to gain assurance over valuation and accuracy in a fully automated process. 

The chains currently supported are as follows:
- Ethereum
- Polygon
- Binance Smart Chain
- Avalanche
- Arbitrum
- Fantom
- Cronos

This tool is a work-in-progress and will eventually have a frontend added to it to make it as user-friendly as possible. 

A tutorial is below on how to use the programme in its current form. 

**Step-by-step instructions**

1. Download the proof-of-balance.exe file
2. Create a CSV with the correct format. A template is provided below:

| Address | Chain |
| --- | --- |
| 0x574977cc4291Be87CabA68a767D542C16FA7c0DD | eth |
| 0x574977cc4291Be87CabA68a767D542C16FA7c0DD | arbitrum |
| 0x574977cc4291Be87CabA68a767D542C16FA7c0DD | bsc |
| 0x574977cc4291Be87CabA68a767D542C16FA7c0DD | polygon |
| 0x574977cc4291Be87CabA68a767D542C16FA7c0DD | fantom |

3. Open the .exe file and select the option you would like (1: Retrieve block number for specific time and chain [coming soon], 2: Perform proof-of-balance)
4. The programme will now ask you to select the CSV file. Choose the file with the template as mentioned in step 2.
5. The programme will now exit and will have saved a file in the same folder where the .exe file is with the name "retrieved-data.csv". 

**Note**
- For now, the only timestamp available is 31/12/2022 23:59:59 UTC. The programme will be amended to be able to take any timestamp from the user.
- This is a work-in-progress and further changes will be made to the programme. The CLI version will remain, but in future there will be a frontend website and a client section where all proof-of-balance history will be displayed. 
