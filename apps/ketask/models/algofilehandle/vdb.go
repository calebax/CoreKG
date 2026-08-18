package algofilehandle

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"unsafe"
)

type VdbEntities struct {
	Data []struct {
		ID         string `json:"__id__"`
		CreatedAt  string `json:"__created_at__"`
		EntityName string `json:"entity_name"`
		Content    string `json:"content"`
		SourceID   string `json:"source_id"`
		FilePath   string `json:"file_path"`
	} `json:"data"`
	EmbeddingMatrix `json:",inline"`
}

type VdbRelationships struct {
	Data []struct {
		ID        string `json:"__id__"`
		CreatedAt string `json:"__created_at__"`
		SrcID     string `json:"src_id"`
		TgtID     string `json:"tgt_id"`
		Content   string `json:"content"`
		SourceID  string `json:"source_id"`
		FilePath  string `json:"file_path"`
	} `json:"data"`
	EmbeddingMatrix `json:",inline"`
}

type VdbChunks struct {
	Data []struct {
		ID        string `json:"__id__"`
		CreatedAt string `json:"__created_at__"`
		Content   string `json:"content"`
		FilePath  string `json:"file_path"`
		Tokens    int    `json:"tokens"`
	} `json:"data"`
	EmbeddingMatrix `json:",inline"`
}

// GenerateVdbObject 解析JSON并生成对应的对象
func GenerateVdbObject(fileByte []byte, targetType any) error {
	err := json.Unmarshal(fileByte, targetType)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return nil
}

// DecodeMatrixFloat32 base64 string to [][]float32
func DecodeMatrixFloat32(matrixBase64 string, dim int) ([][]float32, error) {
	raw, err := base64.StdEncoding.DecodeString(matrixBase64)
	if err != nil {
		return nil, err
	}

	// Read raw bytes into float32 slice
	floatCount := len(raw) / 4 // float32 = 4 bytes
	if floatCount%dim != 0 {
		return nil, fmt.Errorf("data size not divisible by embedding_dim")
	}

	floatData := make([]float32, floatCount)
	for i := 0; i < floatCount; i++ {
		bits := binary.LittleEndian.Uint32(raw[i*4 : (i+1)*4])
		floatData[i] = float32FromBits(bits)
	}

	// Reshape into 2D slice
	rows := floatCount / dim
	matrix := make([][]float32, rows)
	for i := 0; i < rows; i++ {
		matrix[i] = floatData[i*dim : (i+1)*dim]
	}

	return matrix, nil
}

func float32FromBits(bits uint32) float32 {
	return *(*float32)(unsafe.Pointer(&bits))
}

// EmbeddingMatrix represents the structure of the embedding matrix in the VDB file.
type EmbeddingMatrix struct {
	EmbeddingName string `json:"embedding_name"`
	EmbeddingDim  int    `json:"embedding_dim"`
	MatrixBase64  string `json:"matrix"`
}

// GetMatrix decodes the base64-encoded embedding matrix into a 2D slice of float32.
func (e EmbeddingMatrix) GetMatrix() ([][]float32, error) {
	return DecodeMatrixFloat32(e.MatrixBase64, e.EmbeddingDim)
}
