# vtadmin-demo-data-insert

data inserter tool for vtadmin reshard demo

## Usage:

First, build this with `go build`.

In the local example (`vitess.io/vitess/examples/local`):
* run scripts `./101_initial_cluster.sh` up through `./302_new_shards.sh`
* in a second pane, run `./scripts/vtadmin-up.sh`, which will spawn `vtadmin-web` running on `localhost:3000` in your browser.
* navigate to `http://localhost:3000/workflows`, and you should see no workflows running.
* in a third pane, run this tool, with whichever batch-size/threads/sleep-interval params you wish.
    * for example, i was using `./vtadmin-demo-insert-data -threads 16 -batch_size 500 -sleep 100ms`.
* back in your first terminal pane, run `./303_reshard.sh`. this starts the workflow.
* back in your browser, you should see a new workflow, called `cust2cust`. if you click on it you'll see 2 streams, both in Copy phase. Depending on how long you were running the data-inserter, they might be in Copy phase for a long time
* eventually, you will see those streams transition (live!) to `Running` phase. now you can run `VDiff`:
    * `vtctlclient -server "localhost:15999" VDiff -source_cell "zone1" -target_cell "zone1" customer.cust2cust`
    *
```
    Summary for table corder:
        ProcessedRows: 0
        MatchingRows: 0
        MismatchedRows: 0
        ExtraRowsSource: 0
        ExtraRowsTarget: 0
    Summary for table customer:
        ProcessedRows: 2659669
        MatchingRows: 2659669
        MismatchedRows: 0
        ExtraRowsSource: 0
        ExtraRowsTarget: 0
 ```

* Assuming all rows matched, you can finish the migration:
    * `./304_switch_reads.sh`
    * `./305_switch_writes.sh`
* At this point, vtadmin will show the streams in `Stopped` phase (because the reverse vreplication are technically part of a separate workflow from vtadmin's perspective, but they are definitely running)
* That's all folks!
