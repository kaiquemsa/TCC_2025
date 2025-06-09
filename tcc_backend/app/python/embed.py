import sys
import json
import torch
import math
from transformers import AutoTokenizer, AutoModel

model_name = "sentence-transformers/all-MiniLM-L6-v2"
tokenizer = AutoTokenizer.from_pretrained(model_name)
model = AutoModel.from_pretrained(model_name)

def generate_embedding(text):
    inputs = tokenizer(text, return_tensors="pt", truncation=True, padding=True)
    with torch.no_grad():
        outputs = model(**inputs)
    embeddings = outputs.last_hidden_state.mean(dim=1).squeeze().tolist()
    cleaned_embedding = [
        0.0 if (not isinstance(x, float) or math.isnan(x) or math.isinf(x)) else x
        for x in embeddings
    ]
    return cleaned_embedding


if __name__ == "__main__":
    input_text = sys.stdin.read()
    embedding = generate_embedding(input_text.strip())
    print(json.dumps(embedding))  
