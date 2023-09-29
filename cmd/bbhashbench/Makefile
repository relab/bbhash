all: plot

plot:
	python3 plot.py
	cd plot && pdflatex main.tex && open main.pdf
	cd ..

clean:
	rm -rf plot

.PHONY: all plot clean
