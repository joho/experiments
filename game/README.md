# Game scratch pad

Hana Kim's mobile dev with Go talk at gophercon was super awesome and decided I want to play.

## Things I have learned

* the new mobile stuff is really underdocumented and kinda tricky to get going with
* you create a single scene up front, which you add/remove nodes from
* the place you put your game loop stuff appears to be in the `Arranger` on each node
* you need to call render in draw, and while rendering it calls every node's arranger (which is where i've dumped sprite movement logic)
* you cannot set screen dimension up front when devving on your mac

## Licence

MIT for any code, copyright John Barton

Any image assets are copyright original author - from http://graphicriver.net/item/vertical-shooter-kit/9533737
