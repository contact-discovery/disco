# contact-discovery


## Code organization

The directories in this respository are:

- **cf_creator**: Tool that creates Cuckoo filters (CF) and stores them in a file for later use in our protocol
- **cmd**: contains main functions for `provider` and non-mobile `client`
- **cuckoo_sim**: Cuckoo filter simulation. Simulate daily updates of X % to a CF
- **fbs**: Flatbuffers schema 
- **pir**: PIR protocol (optimized [Checklist](https://github.com/dimakogan/checklist))
- **psetggm**: Puncturable Pseudorandom Set from [Checklist](https://github.com/dimakogan/checklist)
- **oprf_c**: calls OPRF functionality from `mobile_psi_cpp`
- **psi**: PSI protocol
- **tests**: Function tests
- **util**: Additional functionality
