# Add tinf source files
TINF_SRC = adler32.c crc32.c tinfgzip.c tinflate.c tinfzlib.c tinf.h

# Update the list of source files
SRC = stager.c aes.c elf_loader.c $(TINF_SRC)

# Configure the download server via environment variables
CFLAGS += -DGOARCH_$(ARCH) -DDOWNLOAD_HOST=\"$(DOWNLOAD_HOST)\" -DDOWNLOAD_PORT=\"$(DOWNLOAD_PORT)\" -DDOWNLOAD_PATH=\"$(DOWNLOAD_PATH)\" -DDOWNLOAD_KEY=\"$(DOWNLOAD_KEY)\"

stager_so:
	${CC} ${CFLAGS} -DOS_LINUX $(SRC) -o stager.so -fPIC -shared -ldl -nostdlib -nodefaultlibs -s
	@echo "Stager shared object compiled successfully at ${PWD}/stager.so"

stager_so_debug:
	${CC} ${CFLAGS} -DDEBUG -DOS_LINUX $(SRC) -o stager.so -pie -fPIC -shared -ldl -nostdlib -nodefaultlibs -g
	@echo "Stager shared object (with debug) compiled successfully at ${PWD}/stager.so"
