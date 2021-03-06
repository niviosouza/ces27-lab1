package main

import (
	"encoding/json"
	"fmt"
	"github.com/pauloaguiar/ces27-lab1/mapreduce"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"unicode"
	"unicode/utf8"
)

const (
	MAP_PATH           = "map/"
	RESULT_PATH        = "result/"
	MAP_BUFFER_SIZE    = 10
	REDUCE_BUFFER_SIZE = 10
)

// fanInData will run a goroutine that reads files crated by splitData and share them with
// the mapreduce framework through the one-way channel. It'll buffer data up to
// MAP_BUFFER_SIZE (files smaller than chunkSize) and resume loading them
// after they are read on the other side of the channle (in the mapreduce package)
func fanInData(numFiles int) chan []byte {
	var (
		err    error
		input  chan []byte
		buffer []byte
	)

	input = make(chan []byte, MAP_BUFFER_SIZE)

	go func() {
		for i := 0; i < numFiles; i++ {
			if buffer, err = ioutil.ReadFile(mapFileName(i)); err != nil {
				close(input)
				log.Fatal(err)
			}

			log.Println("Fanning in file", mapFileName(i))
			input <- buffer
		}
		close(input)
	}()
	return input
}

// fanOutData will run a goroutine that receive data on the one-way channel and will
// proceed to store it in their final destination. The data will come out after the
// reduce phase of the mapreduce model.
func fanOutData() (output chan []mapreduce.KeyValue, done chan bool) {
	var (
		err           error
		file          *os.File
		fileEncoder   *json.Encoder
		reduceCounter int
	)

	output = make(chan []mapreduce.KeyValue, REDUCE_BUFFER_SIZE)
	done = make(chan bool)

	go func() {
		for v := range output {
			log.Println("Fanning out file", resultFileName(reduceCounter))
			if file, err = os.Create(resultFileName(reduceCounter)); err != nil {
				log.Fatal(err)
			}

			fileEncoder = json.NewEncoder(file)

			for _, value := range v {
				fileEncoder.Encode(value)
			}

			file.Close()
			reduceCounter++
		}

		done <- true
	}()

	return output, done
}

// Reads input file and split it into files smaller than chunkSize.
// CUTCUTCUTCUTCUT!
func splitData(fileName string, chunkSize int) (numMapFiles int, err error) {
	// 	When you are reading a file and the end-of-file is found, an error is returned.
	// 	To check for it use the following code:
	// 		if bytesRead, err = file.Read(buffer); err != nil {
	// 			if err == io.EOF {
	// 				// EOF error
	// 			} else {
	//				panic(err)
	//			}
	// 		}
	//
	// 	Use the mapFileName function to generate the name of the files!

	/////////////////////////
	// YOUR CODE GOES HERE //
	/////////////////////////

	// FEITO EM LINUX

	// Abrindo o arquivo
	fileToSplit, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}

	// Levantando informações do arquivo
	stateFile,_ := fileToSplit.Stat()

	numMapFiles = 0

	// Aloncando memória e lendo dados
	dataFile := make([]byte, stateFile.Size())		
	_, err = fileToSplit.Read(dataFile)

	// Função que checa erros
	check(err)

	// Controlando posições
	initialPos := 0
	invalidOffSet := 0

	// Iteração do conteúdo do arquivo
	for byteOffset, runa := range string(dataFile) {
		invalidChar := !unicode.IsLetter(runa) && !unicode.IsNumber(runa)
		// Tamanho em bytes do caracter
		sizeRuna := utf8.RuneLen(runa)

		// Verifica posição final adequada
		if byteOffset - initialPos + sizeRuna <= chunkSize {
			if invalidChar {
				invalidOffSet = byteOffset + sizeRuna
			}

		} else {
			// Gera arquivo para bloco de dados de initialPos até initialPos + counterChunkSize (carater corrente - runa - é inválido) ou de initialPos até invalidOffSet (no caso de o caracter corrente ser válido)

			if invalidChar {

				makeFiles(dataFile, numMapFiles,initialPos,byteOffset)
				initialPos = byteOffset

			} else {

				makeFiles(dataFile, numMapFiles,initialPos,invalidOffSet)
				initialPos = invalidOffSet
			}

			numMapFiles += 1

		}

	}

	fileByteSize := int(stateFile.Size())

	// Resolvendo o fim do arquivo
	if initialPos < fileByteSize {
		makeFiles(dataFile, numMapFiles,initialPos,fileByteSize)
		numMapFiles += 1
	}

	return numMapFiles, nil
}

// Montagem dos arquivos "splitados"
func makeFiles(dataFile []byte, numMapFiles int, initialPos int, finalPos int) {
	nameSmallFile := mapFileName(numMapFiles)
	smallFile,err := os.Create(nameSmallFile)
	check(err)
	sizeFile,err := smallFile.Write(dataFile[initialPos:finalPos])
	check(err)
	fmt.Printf("It has written %d bytes: %q to %s.\n", sizeFile, dataFile[initialPos:finalPos], nameSmallFile)
	defer smallFile.Close()
}

// Função para checagem de erros na manipulação de arquivos
func check(e error) {
    if e != nil {
		log.Fatal(e)
        panic(e)
    }
}

func mapFileName(id int) string {
	return filepath.Join(MAP_PATH, fmt.Sprintf("map-%v", id))
}

func resultFileName(id int) string {
	return filepath.Join(RESULT_PATH, fmt.Sprintf("result-%v", id))
}
