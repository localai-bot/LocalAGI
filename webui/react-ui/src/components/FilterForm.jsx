import React from 'react';
import ConfigForm from './ConfigForm';

/**
 * FilterForm component for configuring an filter
 * Renders filter configuration forms based on field group metadata
 */
const FilterForm = ({ filters = [], onChange, onRemove, onAdd, fieldGroups = [] }) => {
  const handleFilterChange = (index, updatedFilter) => {
    onChange(index, updatedFilter);
  };
  
  return (
    <ConfigForm
      items={filters}
      fieldGroups={fieldGroups}
      onChange={handleFilterChange}
      onRemove={onRemove}
      onAdd={onAdd}
      itemType="filter"
      typeField="name"
      addButtonText="Add Filter"
    />
  );
};

export default FilterForm;
