hash = `../chain33-para-cli send wasm create -x dice -f 10 -n deployDice -d dice -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt`

sleep 15

result = `../chain33 tx query -s ${hash}`

while ${result} = "tx not exist"
do
    sleep 5
    result = `../chain33 tx query -s ${hash}`
done



hash2 = `../chain33-para-cli send wasm call -x startgame -e user.p.para.user.wasm.dice -f 1 -n start -r "{\"deposit\":\"400\"}" -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt`
