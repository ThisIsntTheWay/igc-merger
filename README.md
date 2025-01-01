# IGC merger
Merge 2 IGC files together, such as when an FR has erroneously split up a single flight.  
Only IGC files of an identical FR, pilot and date will be processed.

> [!CAUTION]  
> The resultant IGC file will **fail to pass integrity checks** because its security record cannot be faithfully reconstructed.  
> Said security record relies on secret keys only known to the manufacturer of the flight recorder that has generated your IGC files.  

Nonetheless, a security record gets generated to preserve the IGC file structure.

## Usage
```bash
go run . <file1.igc> <file2.igc>
```

The merged file will be written to the current workin directory as `merged.igc`.

## Building
```bash
# With go
go get
go build .

# Using ko
ko build --local .
```