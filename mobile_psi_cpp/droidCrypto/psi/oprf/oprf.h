// oprf.h
#ifdef __cplusplus
extern "C" {
#endif

//void SayHi();
void doECNR_OPRF(int num_elements, bool first, char* s_addr, int s_port, uint8_t * ptr);
void doGCAES_OPRF(int num_elements, bool first, char* s_addr, int s_port, uint8_t * ptr);
void doGCLowMC_OPRF(int num_element, bool first, char* s_addr, int s_port, uint8_t * ptr);


#ifdef __cplusplus
}
#endif
