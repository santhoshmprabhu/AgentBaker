define retrycmd_if_failure
    retries=$(1); wait_sleep=$(2); timeout=$(3); cmd=$(4); target=$$(basename $$(echo $$cmd)); \
    echo "Running $$cmd with $$retries retries, target is $$target"; \
    for i in $(seq 1 $$retries); do
        timeout $$timeout $$cmd && break || \
        if [ $$i -eq $$retries ]; then
            echo Executed $$target $$i times;
            exit 1
        else
            sleep $$wait_sleep
        fi
    done
    echo Executed $$target $$i times;
endef