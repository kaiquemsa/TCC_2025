import sys, json, math, os
import torch
from transformers import AutoTokenizer, AutoModel

# Desliga paralelismo exagerado e limita threads (economia de RAM/CPU)
os.environ["TOKENIZERS_PARALLELISM"] = "false"
torch.set_num_threads(1)

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
MODEL_DIR = os.path.join(BASE_DIR, "model")

# Carrega tokenizer e modelo apenas do disco
tokenizer = AutoTokenizer.from_pretrained(MODEL_DIR, local_files_only=True)
model = AutoModel.from_pretrained(MODEL_DIR, local_files_only=True)
model.eval()

def generate_embedding(text: str):
    inputs = tokenizer(text, return_tensors="pt", truncation=True, padding=True, max_length=256)
    with torch.no_grad():
        outputs = model(**inputs)
    # mean pooling simples
    emb = outputs.last_hidden_state.mean(dim=1).squeeze().tolist()
    # sanitiza valores
    return [0.0 if (not isinstance(x, float) or math.isnan(x) or math.isinf(x)) else x for x in emb]

if __name__ == "__main__":
    import warnings
    warnings.filterwarnings("ignore")

    # Redireciona stderr pra /dev/null pra evitar logs do torch/transformers
    sys.stderr = open(os.devnull, "w")

    text = sys.stdin.read().strip()
    if not text:
        print("[]", flush=True)
        sys.exit(0)

    try:
        result = generate_embedding(text)
        print(json.dumps(result), flush=True)
    except Exception as e:
        # Em caso de erro, retorna um JSON válido pra evitar crash no Go
        print(json.dumps([]), flush=True)
        sys.exit(1)
