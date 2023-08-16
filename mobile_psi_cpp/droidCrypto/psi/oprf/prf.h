// prf.h
#ifdef __cplusplus
extern "C" {
#endif


void getECNR_PRF(int num_elements, bool first, int num_threads, uint8_t* ptr);
void getGCAES_PRF(int num_elements, bool first, int num_threads, uint8_t* ptr);
void getGCLowMC_PRF(int num_elements, bool first, int num_threads, uint8_t* ptr);

#ifdef __cplusplus
}
#endif
